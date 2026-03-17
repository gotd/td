package telegram

import (
	"context"
	"crypto/rsa"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
)

func newCDNPoolTestClient() *Client {
	c := &Client{
		log: zap.NewNop(),
	}
	c.init()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.cfg.Store(tg.Config{
		DCOptions: []tg.DCOption{{
			ID:        203,
			IPAddress: "127.0.0.1",
			Port:      443,
			CDN:       true,
		}},
	})
	// Skip network in tests: prefill CDN cache and keep at least one bundled key.
	baseKey := PublicKey{RSA: &rsa.PublicKey{N: big.NewInt(251), E: 65537}}
	c.opts.PublicKeys = []PublicKey{baseKey}
	c.cdnKeysSet = true
	c.cdnKeys = []PublicKey{baseKey}
	c.cdnKeysByDC = map[int][]PublicKey{
		203: {baseKey},
	}

	return c
}

func unwrapCDNHandle(t *testing.T, inv CloseInvoker) *cdnPoolHandle {
	t.Helper()

	h, ok := inv.(*cdnPoolHandle)
	require.True(t, ok)
	return h
}

func cachedCDNHandleForTest(m *cdnPoolManager, conn CloseInvoker) CloseInvoker {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.cachedHandleLocked(conn)
}

func closeCachedCDNPools(c *Client) {
	c.cdnPools.mux.Lock()
	defer c.cdnPools.mux.Unlock()

	for _, pools := range c.cdnPools.conns {
		for _, cached := range pools {
			_ = cached.conn.Close()
		}
	}
}

type countingCloseInvoker struct {
	closed int
}

func (s *countingCloseInvoker) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *countingCloseInvoker) Close() error {
	s.closed++
	return nil
}

type signalCloseInvoker struct {
	once sync.Once
	ch   chan struct{}
}

func (*signalCloseInvoker) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *signalCloseInvoker) Close() error {
	s.once.Do(func() {
		close(s.ch)
	})
	return nil
}

type idlePoolConn struct {
	ready chan struct{}
}

func newIdlePoolConn() *idlePoolConn {
	ready := make(chan struct{})
	close(ready)
	return &idlePoolConn{ready: ready}
}

func (c *idlePoolConn) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (*idlePoolConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (*idlePoolConn) Ping(context.Context) error {
	return nil
}

func (c *idlePoolConn) Ready() <-chan struct{} {
	return c.ready
}

func TestClientCDNPoolCacheRespectsMax(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()
	defer closeCachedCDNPools(c)

	first, err := c.CDN(context.Background(), 203, 1)
	require.NoError(t, err)
	second, err := c.CDN(context.Background(), 203, 8)
	require.NoError(t, err)
	require.NotSame(t, first, second)

	reused, err := c.CDN(context.Background(), 203, 2)
	require.NoError(t, err)
	require.NotSame(t, second, reused)

	firstShared := unwrapCDNHandle(t, first).conn
	secondShared := unwrapCDNHandle(t, second).conn
	reusedShared := unwrapCDNHandle(t, reused).conn
	require.NotSame(t, firstShared, secondShared)
	require.Same(t, secondShared, reusedShared)

	c.cdnPools.mux.Lock()
	pools := append([]cachedCDNPool(nil), c.cdnPools.conns[203]...)
	c.cdnPools.mux.Unlock()
	require.Len(t, pools, 2)
}

func TestNormalizeCDNPoolMax(t *testing.T) {
	require.EqualValues(t, 0, normalizeCDNPoolMax(0))
	require.EqualValues(t, 1, normalizeCDNPoolMax(1))
	require.EqualValues(t, 2, normalizeCDNPoolMax(2))
	require.EqualValues(t, 4, normalizeCDNPoolMax(3))
	require.EqualValues(t, 4, normalizeCDNPoolMax(4))
	require.EqualValues(t, 8, normalizeCDNPoolMax(5))
}

func TestClientCDNPoolCacheNormalizesNearbyMax(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()
	defer closeCachedCDNPools(c)

	first, err := c.CDN(context.Background(), 203, 3)
	require.NoError(t, err)
	second, err := c.CDN(context.Background(), 203, 4)
	require.NoError(t, err)
	require.NotSame(t, first, second)

	third, err := c.CDN(context.Background(), 203, 5)
	require.NoError(t, err)
	require.NotSame(t, second, third)

	firstShared := unwrapCDNHandle(t, first).conn
	secondShared := unwrapCDNHandle(t, second).conn
	thirdShared := unwrapCDNHandle(t, third).conn
	require.Same(t, firstShared, secondShared)
	require.NotSame(t, secondShared, thirdShared)

	c.cdnPools.mux.Lock()
	pools := append([]cachedCDNPool(nil), c.cdnPools.conns[203]...)
	c.cdnPools.mux.Unlock()
	require.Len(t, pools, 2)
}

func TestClientCDNPoolCloseKeepsCacheForReuse(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()
	defer closeCachedCDNPools(c)

	first, err := c.CDN(context.Background(), 203, 4)
	require.NoError(t, err)
	firstShared := unwrapCDNHandle(t, first).conn
	require.NoError(t, first.Close())

	c.cdnPools.mux.Lock()
	poolsAfterClose := append([]cachedCDNPool(nil), c.cdnPools.conns[203]...)
	refsAfterClose := c.cdnPools.refs[firstShared]
	c.cdnPools.mux.Unlock()
	require.Len(t, poolsAfterClose, 1)
	require.EqualValues(t, 1, refsAfterClose)

	second, err := c.CDN(context.Background(), 203, 4)
	require.NoError(t, err)
	secondShared := unwrapCDNHandle(t, second).conn
	require.Same(t, firstShared, secondShared)
	require.NoError(t, second.Close())
}

func TestClientCDNPoolHandleDoubleClose(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()
	defer closeCachedCDNPools(c)

	h, err := c.CDN(context.Background(), 203, 4)
	require.NoError(t, err)
	require.NoError(t, h.Close())
	require.ErrorIs(t, h.Close(), errCDNPoolHandleDouble)
}

func TestClientCDNPoolHandleCloseWaitsForLastHandle(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()
	defer closeCachedCDNPools(c)

	const dcID = 203
	shared := &countingCloseInvoker{}
	c.cdnPools.conns[dcID] = []cachedCDNPool{{
		conn: shared,
		max:  4,
	}}

	first := cachedCDNHandleForTest(&c.cdnPools, shared)
	second := cachedCDNHandleForTest(&c.cdnPools, shared)

	require.NoError(t, first.Close())
	require.Equal(t, 0, shared.closed, "shared pool must stay alive while second handle is active")

	c.cdnPools.mux.Lock()
	_, ok := pickCDNPool(c.cdnPools.conns[dcID], 1)
	c.cdnPools.mux.Unlock()
	require.True(t, ok, "cache entry must remain until last handle closes")

	require.NoError(t, second.Close())
	require.Equal(t, 0, shared.closed, "underlying pool must remain open in cache after last handle close")

	c.cdnPools.mux.Lock()
	refs := c.cdnPools.refs[shared]
	c.cdnPools.mux.Unlock()
	require.EqualValues(t, 1, refs, "cache owner ref should remain for reuse")
}

func TestCDNPoolManagerDrainIncludesPendingCloseQueue(t *testing.T) {
	m := newCDNPoolManager()

	connA := &countingCloseInvoker{}
	connB := &countingCloseInvoker{}
	connC := &countingCloseInvoker{}
	m.conns[203] = []cachedCDNPool{
		{conn: connA, max: 1},
		{conn: connB, max: 2},
	}
	// connB is duplicated intentionally: drain() must deduplicate results.
	m.closeQueue = []CloseInvoker{connB, connC}

	drained := m.drain()
	require.Len(t, drained, 3)
	require.Empty(t, m.conns)
	require.Empty(t, m.refs)
	require.Empty(t, m.closeQueue)
}

func TestCDNPoolManagerQueueSaturationDoesNotSpawnDetachedClose(t *testing.T) {
	m := newCDNPoolManager()
	queued := &signalCloseInvoker{ch: make(chan struct{})}

	m.mux.Lock()
	m.closeWorkers = maxCDNCloseWorkers
	m.closeBusy = maxCDNCloseWorkers
	m.closeQueue = make([]CloseInvoker, maxCDNCloseQueue)
	for i := range m.closeQueue {
		m.closeQueue[i] = &countingCloseInvoker{}
	}
	m.enqueueCloseLocked([]CloseInvoker{queued})
	require.Len(t, m.closeQueue, maxCDNCloseQueue)
	m.mux.Unlock()

	select {
	case <-queued.ch:
		t.Fatal("unexpected detached close while workers are saturated")
	case <-time.After(100 * time.Millisecond):
	}

	m.mux.Lock()
	require.Len(t, m.closeQueue, maxCDNCloseQueue)
	m.mux.Unlock()
}

type blockingSignalInvoker struct {
	started chan struct{}
	unlock  chan struct{}
	done    chan struct{}
	once    sync.Once
}

func (*blockingSignalInvoker) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *blockingSignalInvoker) Close() error {
	s.once.Do(func() {
		close(s.started)
	})
	<-s.unlock
	close(s.done)
	return nil
}

func TestCDNPoolManagerQueueOverflowEventuallyClosesPending(t *testing.T) {
	m := newCDNPoolManager()

	blocked := &blockingSignalInvoker{
		started: make(chan struct{}),
		unlock:  make(chan struct{}),
		done:    make(chan struct{}),
	}
	overflow := &signalCloseInvoker{ch: make(chan struct{})}

	m.mux.Lock()
	m.closeWorkers = 1
	m.closeBusy = 1
	m.closeQueue = make([]CloseInvoker, 0, maxCDNCloseQueue)
	m.closeQueue = append(m.closeQueue, blocked)
	for i := 1; i < maxCDNCloseQueue; i++ {
		m.closeQueue = append(m.closeQueue, &countingCloseInvoker{})
	}
	m.closing[blocked] = true
	m.enqueueCloseLocked([]CloseInvoker{overflow})
	require.Len(t, m.closeQueue, maxCDNCloseQueue)
	m.mux.Unlock()

	// Free one worker slot and start workers manually.
	m.mux.Lock()
	m.closeBusy = 0
	m.closeWorkers = 0
	m.mux.Unlock()
	go m.runCloseWorker()

	select {
	case <-blocked.started:
	case <-time.After(time.Second):
		t.Fatal("expected queued blocked close to start")
	}

	close(blocked.unlock)

	select {
	case <-overflow.ch:
	case <-time.After(time.Second):
		t.Fatal("expected overflow pending close to be processed after queue drains")
	}
}

func TestClientCDNUsesDCSpecificKeysOverBase(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	baseKey := PublicKey{RSA: &rsa.PublicKey{N: big.NewInt(257), E: 65537}}
	cdnKey := PublicKey{RSA: &rsa.PublicKey{N: big.NewInt(263), E: 65537}}

	c.opts.PublicKeys = []PublicKey{baseKey}
	c.cdnKeysSet = true
	c.cdnKeys = []PublicKey{cdnKey}
	c.cdnKeysByDC = map[int][]PublicKey{
		203: {cdnKey},
	}

	captured := make(chan []PublicKey, 1)
	c.create = func(
		_ mtproto.Dialer,
		mode manager.ConnMode,
		_ int,
		opts mtproto.Options,
		_ manager.ConnOptions,
	) pool.Conn {
		if mode == manager.ConnModeCDN {
			captured <- append([]PublicKey(nil), opts.PublicKeys...)
		}
		return newIdlePoolConn()
	}

	inv, err := c.CDN(context.Background(), 203, 1)
	require.NoError(t, err)
	require.NotNil(t, inv)
	defer func() {
		require.NoError(t, inv.Close())
	}()
	require.NoError(t, inv.Invoke(context.Background(), nil, nil))

	got := <-captured
	require.Equal(t, []PublicKey{cdnKey, baseKey}, got)
}

func TestClientCDNWithoutDCSpecificKeysFailsFast(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	baseKey := PublicKey{RSA: &rsa.PublicKey{N: big.NewInt(269), E: 65537}}
	c.opts.PublicKeys = []PublicKey{baseKey}
	c.cdnKeysSet = true
	c.cdnKeys = nil
	c.cdnKeysByDC = map[int][]PublicKey{}
	c.tg = tg.NewClient(InvokeFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return errors.New("cdn config unavailable")
	}))

	var calls atomic.Int32
	c.create = func(
		_ mtproto.Dialer,
		_ manager.ConnMode,
		_ int,
		_ mtproto.Options,
		_ manager.ConnOptions,
	) pool.Conn {
		calls.Add(1)
		return newIdlePoolConn()
	}

	inv, err := c.CDN(context.Background(), 203, 1)
	require.Error(t, err)
	require.ErrorContains(t, err, "fetch CDN public keys for DC 203")
	require.Nil(t, inv)
	require.Zero(t, calls.Load())
}

func TestClientCDNWithoutAnyKeysFailsFast(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	c.opts.PublicKeys = nil
	c.cdnKeysSet = true
	c.cdnKeys = nil
	c.cdnKeysByDC = map[int][]PublicKey{}
	c.tg = tg.NewClient(InvokeFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return errors.New("cdn config unavailable")
	}))

	var calls atomic.Int32
	c.create = func(
		_ mtproto.Dialer,
		_ manager.ConnMode,
		_ int,
		_ mtproto.Options,
		_ manager.ConnOptions,
	) pool.Conn {
		calls.Add(1)
		return newIdlePoolConn()
	}

	inv, err := c.CDN(context.Background(), 203, 1)
	require.Error(t, err)
	require.Nil(t, inv)
	require.Zero(t, calls.Load())
}

func TestClientCDNFetchDCKeysErrorFailsFast(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	baseKey := PublicKey{RSA: &rsa.PublicKey{N: big.NewInt(271), E: 65537}}
	c.opts.PublicKeys = []PublicKey{baseKey}
	c.cdnKeysSet = false
	c.cdnKeys = nil
	c.cdnKeysByDC = nil
	c.tg = tg.NewClient(InvokeFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return errors.New("cdn config unavailable")
	}))

	var calls atomic.Int32
	c.create = func(
		_ mtproto.Dialer,
		_ manager.ConnMode,
		_ int,
		_ mtproto.Options,
		_ manager.ConnOptions,
	) pool.Conn {
		calls.Add(1)
		return newIdlePoolConn()
	}

	inv, err := c.CDN(context.Background(), 203, 1)
	require.Error(t, err)
	require.ErrorContains(t, err, "fetch CDN public keys for DC 203")
	require.Nil(t, inv)
	require.Zero(t, calls.Load())
}
