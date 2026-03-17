package telegram

import (
	"context"
	"math/bits"
	"sync"
	"sync/atomic"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

type cachedCDNPool struct {
	conn CloseInvoker
	// max is normalized bucket size used for reuse matching.
	max int64
}

var (
	errCDNPoolHandleClosed = errors.New("CDN pool handle is closed")
	errCDNPoolHandleDouble = errors.New("CDN pool handle already closed")
)

// cdnPoolHandle is a per-call wrapper around shared cached CDN pool.
// Close() releases only this borrowed handle; underlying pool is managed by
// client cache lifecycle (fingerprint invalidation or client shutdown).
type cdnPoolHandle struct {
	manager *cdnPoolManager
	conn    CloseInvoker
	closed  atomic.Bool
}

var _ CloseInvoker = (*cdnPoolHandle)(nil)

func (h *cdnPoolHandle) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if h.closed.Load() {
		return errCDNPoolHandleClosed
	}
	return h.conn.Invoke(ctx, input, output)
}

func (h *cdnPoolHandle) Close() error {
	if h.closed.Swap(true) {
		return errCDNPoolHandleDouble
	}
	if !h.manager.releaseCachedHandle(h.conn) {
		return nil
	}
	return h.conn.Close()
}

type cdnPoolManager struct {
	mux sync.Mutex

	conns map[int][]cachedCDNPool
	// refs tracks active handle references for shared CDN pools.
	refs map[CloseInvoker]int
	// closing tracks pools already known to close pipeline to avoid duplicate
	// queue entries and duplicate Close() calls.
	//
	// Value denotes whether conn is currently queued for worker processing.
	closing map[CloseInvoker]bool

	// closeQueue contains stale CDN pools waiting for async close.
	// Close() may block on unstable network/proxy, so queue is processed by
	// bounded worker count.
	closeQueue   []CloseInvoker
	closePending []CloseInvoker
	closeWorkers int
	closeBusy    int
}

const (
	maxCDNCloseWorkers = 4
	// Historical sizing hint for close backlog in tests and heuristics.
	// Queue growth is controlled by stale-pool production rate and de-dup via
	// closing map, while close fan-out remains bounded by workers.
	maxCDNCloseQueue = 256
)

func newCDNPoolManager() cdnPoolManager {
	return cdnPoolManager{
		conns:   map[int][]cachedCDNPool{},
		refs:    map[CloseInvoker]int{},
		closing: map[CloseInvoker]bool{},
	}
}

func (p cachedCDNPool) covers(need int64) bool {
	// pool max < 1 means unlimited in pool package.
	if p.max < 1 {
		return true
	}
	// Requested max < 1 means unlimited, finite pool does not satisfy it.
	if need < 1 {
		return false
	}
	return p.max >= need
}

func pickCDNPool(pools []cachedCDNPool, need int64) (CloseInvoker, bool) {
	// Pick the smallest pool that still covers requested capacity.
	best := -1
	for i, p := range pools {
		if !p.covers(need) {
			continue
		}
		if best == -1 {
			best = i
			continue
		}
		// Prefer tighter finite limit to avoid over-allocating.
		if pools[best].max < 1 {
			best = i
			continue
		}
		if p.max > 0 && p.max < pools[best].max {
			best = i
		}
	}
	if best == -1 {
		return nil, false
	}
	return pools[best].conn, true
}

func (m *cdnPoolManager) cachedHandleLocked(conn CloseInvoker) CloseInvoker {
	refs, ok := m.refs[conn]
	if !ok {
		// Keep one cache-owner reference so pool can be reused between
		// sequential downloads.
		refs = 1
	}
	m.refs[conn] = refs + 1

	return &cdnPoolHandle{
		manager: m,
		conn:    conn,
	}
}

func (m *cdnPoolManager) releaseCachedHandle(conn CloseInvoker) bool {
	m.mux.Lock()
	defer m.mux.Unlock()

	refs, ok := m.refs[conn]
	if !ok || refs < 1 {
		// Connection is already evicted/closed by another path.
		return false
	}
	refs--
	if refs == 0 {
		delete(m.refs, conn)
		return true
	}
	m.refs[conn] = refs
	return false
}

func (m *cdnPoolManager) acquire(dc int, need int64) (CloseInvoker, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	cached, ok := pickCDNPool(m.conns[dc], need)
	if !ok {
		return nil, false
	}
	return m.cachedHandleLocked(cached), true
}

func (m *cdnPoolManager) publishOrAcquire(dc int, need int64, created CloseInvoker) (CloseInvoker, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if existing, ok := pickCDNPool(m.conns[dc], need); ok {
		return m.cachedHandleLocked(existing), true
	}
	m.conns[dc] = append(m.conns[dc], cachedCDNPool{
		conn: created,
		max:  need,
	})
	return m.cachedHandleLocked(created), false
}

func (m *cdnPoolManager) drain() []CloseInvoker {
	m.mux.Lock()
	defer m.mux.Unlock()

	seen := map[CloseInvoker]struct{}{}
	cdnConns := make([]CloseInvoker, 0, len(m.conns)+len(m.closeQueue)+len(m.closePending))
	add := func(conn CloseInvoker) {
		if conn == nil {
			return
		}
		if _, ok := seen[conn]; ok {
			return
		}
		seen[conn] = struct{}{}
		cdnConns = append(cdnConns, conn)
	}
	for _, pools := range m.conns {
		for _, cached := range pools {
			add(cached.conn)
		}
	}
	for _, conn := range m.closeQueue {
		add(conn)
	}
	for _, conn := range m.closePending {
		add(conn)
	}
	m.conns = map[int][]cachedCDNPool{}
	m.refs = map[CloseInvoker]int{}
	m.closing = map[CloseInvoker]bool{}
	m.closeQueue = nil
	m.closePending = nil
	return cdnConns
}

func (m *cdnPoolManager) refillCloseQueueLocked() {
	for len(m.closeQueue) < maxCDNCloseQueue && len(m.closePending) > 0 {
		conn := m.closePending[0]
		m.closePending[0] = nil
		m.closePending = m.closePending[1:]
		if conn == nil {
			continue
		}

		queued, ok := m.closing[conn]
		if !ok || queued {
			// Already closed/queued by another path.
			continue
		}
		m.closing[conn] = true
		m.closeQueue = append(m.closeQueue, conn)
	}
}

func (m *cdnPoolManager) enqueueCloseLocked(stale []CloseInvoker) {
	if len(stale) == 0 {
		return
	}

	for _, conn := range stale {
		if conn == nil {
			continue
		}
		if _, ok := m.closing[conn]; ok {
			continue
		}
		if len(m.closeQueue) < maxCDNCloseQueue {
			m.closing[conn] = true
			m.closeQueue = append(m.closeQueue, conn)
			continue
		}

		// Queue is saturated, keep pending task deduplicated and promote when
		// worker frees queue slots.
		m.closing[conn] = false
		m.closePending = append(m.closePending, conn)
	}

	m.refillCloseQueueLocked()

	// Start enough workers to avoid head-of-line blocking on slow Close(),
	// but keep fan-out bounded.
	for m.closeWorkers < maxCDNCloseWorkers {
		available := m.closeWorkers - m.closeBusy
		if available >= len(m.closeQueue) {
			break
		}
		m.closeWorkers++
		go m.runCloseWorker()
	}
}

func (m *cdnPoolManager) runCloseWorker() {
	for {
		m.mux.Lock()
		if len(m.closeQueue) == 0 {
			m.closeWorkers--
			m.mux.Unlock()
			return
		}
		conn := m.closeQueue[0]
		m.closeQueue[0] = nil
		m.closeQueue = m.closeQueue[1:]
		m.closeBusy++
		m.mux.Unlock()

		_ = conn.Close()

		m.mux.Lock()
		delete(m.closing, conn)
		m.closeBusy--
		m.refillCloseQueueLocked()
		m.mux.Unlock()
	}
}

func (m *cdnPoolManager) invalidateDC(dcID int) {
	m.mux.Lock()
	stale := append([]cachedCDNPool(nil), m.conns[dcID]...)
	for _, cached := range stale {
		delete(m.refs, cached.conn)
	}
	delete(m.conns, dcID)

	toClose := make([]CloseInvoker, 0, len(stale))
	for _, cached := range stale {
		toClose = append(toClose, cached.conn)
	}
	m.enqueueCloseLocked(toClose)
	m.mux.Unlock()
}

func normalizeCDNPoolMax(max int64) int64 {
	// Keep unlimited pools as-is.
	if max < 1 {
		return max
	}
	// Collapse close finite values into power-of-two buckets to cap the number
	// of cached CDN pools per DC.
	if max < 2 {
		return max
	}
	shift := bits.Len64(uint64(max - 1))
	// Guard signed overflow for extremely large values.
	if shift >= 63 {
		return max
	}
	return int64(1) << shift
}
