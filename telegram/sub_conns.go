package telegram

import (
	"context"
	"sync"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/auth"
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
	// kind is the pool kind for metrics and logs.
	kind string
}

func newDCPoolCache(max int64, kind string) *dcPoolCache {
	return &dcPoolCache{
		conns:   map[int]CloseInvoker{},
		pending: map[int]*subConnCall{},
		max:     max,
		kind:    kind,
	}
}

// len returns the number of cached pools.
func (p *dcPoolCache) len() int {
	p.mux.Lock()
	defer p.mux.Unlock()
	return len(p.conns)
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

// evict removes conn from cache if it is still the cached pool of dc and
// closes it.
func (p *dcPoolCache) evict(dc int, conn CloseInvoker) {
	p.mux.Lock()
	cur, ok := p.conns[dc]
	if !ok || cur != conn {
		p.mux.Unlock()
		return
	}
	delete(p.conns, dc)
	p.mux.Unlock()
	_ = conn.Close()
}

// errClientNotRunning is returned when DC pool is requested before Run.
var errClientNotRunning = errors.New("client is not running")

func (c *Client) invokeSub(ctx context.Context, dc int, input bin.Encoder, output bin.Decoder) error {
	return c.invokeCachedDC(ctx, c.subConns, dc, input, output)
}

// invokeCachedDC invokes request on cached data pool of dc.
//
// If the request fails as unauthorized, the pool may be cached with a failed
// authorization transfer, e.g. created before sign in completed: it is
// dropped and the request is retried once on a new pool, which transfers
// authorization again.
func (c *Client) invokeCachedDC(
	ctx context.Context,
	cache *dcPoolCache,
	dc int,
	input bin.Encoder,
	output bin.Decoder,
) error {
	for attempt := 0; ; attempt++ {
		conn, err := c.cachedDC(ctx, cache, dc)
		if err != nil {
			return errors.Wrapf(err, "create connection to DC %d", dc)
		}
		err = conn.Invoke(ctx, input, output)
		if attempt == 0 && auth.IsUnauthorized(err) {
			c.log.Warn(ctx, "Unauthorized on cached pool, recreating it",
				log.Int("dc_id", dc),
				log.String("kind", cache.kind),
				log.Error(err),
			)
			c.meters().poolEvicted.Add(ctx, 1, dcAttrs(dc,
				attribute.String("kind", cache.kind),
				attribute.String("reason", "unauthorized"),
			))
			cache.evict(dc, conn)
			continue
		}
		return err
	}
}

// cachedDC returns cached data pool to given DC or creates new one.
//
// Creation (dial, key exchange and authorization transfer) is done once per DC
// without holding cache lock, so other DCs and calls to already created pools
// are not blocked. Creation is bounded by migration timeout and is not
// canceled when the caller that started it gives up. Failed creation is not
// cached.
func (c *Client) cachedDC(ctx context.Context, cache *dcPoolCache, dc int) (CloseInvoker, error) {
	if c.ctx == nil {
		// Pools are bound to client context, which is set by Run.
		return nil, errClientNotRunning
	}
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
	cctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()
	stop := context.AfterFunc(clientCtx, cancel)
	defer stop()
	if c.migrationTimeout > 0 {
		var cancelTimeout context.CancelFunc
		cctx, cancelTimeout = context.WithTimeout(cctx, c.migrationTimeout)
		defer cancelTimeout()
	}

	start := c.now()
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
	// Waiters are released only after the result is recorded and logged, so
	// they observe it, and nothing of this call outlives them.
	defer close(call.done)

	// Creation is detached from the caller's context: only client shutdown
	// cancels it, which is not a failure of the DC.
	canceled := err != nil && clientCtx.Err() != nil
	result := "ok"
	switch {
	case canceled:
		result = "canceled"
	case err != nil:
		result = "error"
	}
	elapsed := c.now().Sub(start)
	m := c.meters()
	m.poolCreated.Add(ctx, 1, dcAttrs(dc,
		attribute.String("kind", cache.kind),
		attribute.String("result", result),
	))
	m.poolDuration.Record(ctx, elapsed.Seconds(), metric.WithAttributes(
		attribute.String("kind", cache.kind),
		attribute.String("result", result),
	))

	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		logFailure := c.log.Warn
		if canceled {
			logFailure = c.log.Debug
		}
		logFailure(ctx, "Pool creation failed",
			log.Int("dc_id", dc),
			log.String("kind", cache.kind),
			log.Int64("max", cache.max),
			log.Duration("duration", elapsed),
			log.Error(err),
		)
		return
	}
	c.log.Debug(ctx, "Pool created",
		log.Int("dc_id", dc),
		log.String("kind", cache.kind),
		log.Int64("max", cache.max),
		log.Duration("duration", elapsed),
	)
}
