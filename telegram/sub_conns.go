package telegram

import (
	"context"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/manager"
)

// subConnCall is an in-flight creation of a pool to some DC.
type subConnCall struct {
	done chan struct{}
	conn CloseInvoker
	err  error
}

// dcPoolCache caches data pools per DC.
type dcPoolCache struct {
	mux     sync.Mutex
	conns   map[int]CloseInvoker
	pending map[int]*subConnCall
	// max is the pool size.
	max int64
}

func newDCPoolCache(max int64) *dcPoolCache {
	return &dcPoolCache{
		conns:   map[int]CloseInvoker{},
		pending: map[int]*subConnCall{},
		max:     max,
	}
}

// drain removes and returns all cached pools.
func (p *dcPoolCache) drain() []CloseInvoker {
	p.mux.Lock()
	defer p.mux.Unlock()
	conns := make([]CloseInvoker, 0, len(p.conns))
	for _, conn := range p.conns {
		conns = append(conns, conn)
	}
	p.conns = map[int]CloseInvoker{}
	return conns
}

func (c *Client) invokeSub(ctx context.Context, dc int, input bin.Encoder, output bin.Decoder) error {
	conn, err := c.cachedDC(ctx, c.subConns, dc)
	if err != nil {
		return errors.Wrapf(err, "create connection to DC %d", dc)
	}
	return conn.Invoke(ctx, input, output)
}

// cachedDC returns cached data pool to given DC or creates new one.
//
// Creation (dial, key exchange and authorization transfer) is done once per DC
// without holding cache lock, so other DCs and calls to already created pools
// are not blocked. Creation is bounded by migration timeout and is not
// canceled when the caller that started it gives up. Failed creation is not
// cached.
func (c *Client) cachedDC(ctx context.Context, cache *dcPoolCache, dc int) (CloseInvoker, error) {
	cache.mux.Lock()
	if conn, ok := cache.conns[dc]; ok {
		cache.mux.Unlock()
		return conn, nil
	}
	call, ok := cache.pending[dc]
	if !ok {
		call = &subConnCall{done: make(chan struct{})}
		cache.pending[dc] = call
		go c.createCachedDC(ctx, cache, dc, call)
	}
	cache.mux.Unlock()

	select {
	case <-call.done:
		return call.conn, call.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) createCachedDC(ctx context.Context, cache *dcPoolCache, dc int, call *subConnCall) {
	// Keep caller values (tracing, logging), but not its cancellation:
	// creation is shared by all callers waiting for this DC.
	clientCtx := c.ctx
	if clientCtx == nil {
		clientCtx = context.Background()
	}
	cctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()
	stop := context.AfterFunc(clientCtx, cancel)
	defer stop()
	if c.migrationTimeout > 0 {
		var cancelTimeout context.CancelFunc
		cctx, cancelTimeout = context.WithTimeout(cctx, c.migrationTimeout)
		defer cancelTimeout()
	}

	conn, err := c.dc(cctx, dc, cache.max, c.primaryDC(dc), manager.ConnModeData)
	if err != nil && clientCtx.Err() != nil {
		err = errors.Wrap(clientCtx.Err(), "client closed")
	}

	cache.mux.Lock()
	delete(cache.pending, dc)
	if err == nil {
		if clientCtx.Err() != nil {
			// Client is closing (or closed) and has already drained the cache.
			err = errors.Wrap(clientCtx.Err(), "client closed")
		} else {
			cache.conns[dc] = conn
		}
	}
	call.conn, call.err = conn, err
	if err != nil {
		call.conn = nil
	}
	cache.mux.Unlock()
	close(call.done)

	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		c.log.Debug(ctx, "Pool creation failed",
			log.Int("dc_id", dc),
			log.Int64("max", cache.max),
			log.Error(err),
		)
	}
}
