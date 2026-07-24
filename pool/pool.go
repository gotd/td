// Package pool contains Telegram connections pool.
package pool

import (
	"context"
	"sync"

	"github.com/go-faster/errors"
	"github.com/gotd/log"
	"go.uber.org/atomic"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdsync"
)

// DC represents connection pool to one data center.
type DC struct {
	id int

	// Connection constructor.
	newConn func() Conn

	// Wrappers for external world, like logs or PRNG.
	log log.Helper // immutable

	// DC context. Will be canceled by Run on exit.
	ctx    context.Context    // immutable
	cancel context.CancelFunc // immutable

	// Connections supervisor.
	grp *tdsync.Supervisor
	// Free connections.
	free []*poolConn
	// Total connections.
	total int64
	// Connection id monotonic counter.
	nextConn atomic.Int64
	freeReq  *reqMap
	// DC mutex.
	mu sync.Mutex

	// Limit of connections.
	max int64 // immutable

	// Retry a request whose transport send failed on another connection.
	retryOnWriteFailed bool // immutable

	// Signal connection for cases when all connections are dead, but all requests waiting for
	// free connection in 3rd acquire case.
	stuck *tdsync.ResetReady

	closed atomic.Bool
}

// NewDC creates new uninitialized DC.
func NewDC(ctx context.Context, id int, newConn func() Conn, opts DCOptions) *DC {
	ctx, cancel := context.WithCancel(ctx)

	opts.setDefaults()
	return &DC{
		id:      id,
		newConn: newConn,
		log:     log.For(opts.Logger),
		ctx:     ctx,
		cancel:  cancel,
		grp:     tdsync.NewSupervisor(ctx),
		freeReq: newReqMap(),
		max:     opts.MaxOpenConnections,
		stuck:   tdsync.NewResetReady(),

		retryOnWriteFailed: opts.RetryOnWriteFailed,
	}
}

func (c *DC) createConnection(id int64) *poolConn {
	conn := &poolConn{
		Conn:    c.newConn(),
		id:      id,
		dc:      c,
		deleted: atomic.NewBool(false),
		dead:    tdsync.NewReady(),
	}

	c.grp.Go(func(groupCtx context.Context) (err error) {
		defer c.dead(conn, err)
		return conn.Run(groupCtx)
	})

	return conn
}

func (c *DC) dead(r *poolConn, deadErr error) {
	ctx := c.ctx
	if r.deleted.Swap(true) {
		return // Already deleted.
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.total--
	remaining := c.total
	if remaining < 0 {
		panic("unreachable: remaining can't be less than zero")
	}

	idx := -1
	for i, conn := range c.free {
		// Search connection by pointer.
		if conn.id == r.id {
			idx = i
		}
	}

	if idx >= 0 {
		// Delete by index from slice tricks.
		copy(c.free[idx:], c.free[idx+1:])
		// Delete reference to prevent resource leaking.
		c.free[len(c.free)-1] = nil
		c.free = c.free[:len(c.free)-1]
	}

	r.dead.Signal()
	c.stuck.Reset()

	c.log.Debug(ctx, "Connection died",
		log.Int64("remaining", remaining),
		log.Int64("conn_id", r.id),
		log.Error(deadErr),
	)
}

func (c *DC) pop() (r *poolConn, ok bool) {
	l := len(c.free)
	if l > 0 {
		r, c.free = c.free[l-1], c.free[:l-1]

		return r, true
	}

	return
}

func (c *DC) release(r *poolConn) {
	if r == nil {
		return
	}
	ctx := c.ctx

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.freeReq.transfer(r) {
		c.log.Debug(ctx, "Transfer connection to requester", log.Int64("conn_id", r.id))
		return
	}
	c.log.Debug(ctx, "Connection released", log.Int64("conn_id", r.id))
	c.free = append(c.free, r)
}

var errDCIsClosed = errors.New("DC is closed")

func (c *DC) acquire(ctx context.Context) (r *poolConn, err error) { // nolint:gocyclo
retry:
	c.mu.Lock()
	// 1st case: have free connections.
	if r, ok := c.pop(); ok {
		c.mu.Unlock()
		select {
		case <-r.Dead():
			c.dead(r, nil)
			goto retry
		default:
		}
		c.log.Debug(ctx, "Re-using free connection", log.Int64("conn_id", r.id))
		return r, nil
	}

	// 2nd case: no free connections, but can create one.
	// c.max < 1 means unlimited
	if c.max < 1 || c.total < c.max {
		c.total++
		c.mu.Unlock()

		id := c.nextConn.Inc()
		c.log.Debug(ctx, "Creating new connection",
			log.Int64("conn_id", id),
		)
		conn := c.createConnection(id)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.ctx.Done():
			return nil, errors.Wrap(c.ctx.Err(), "DC closed")
		case <-conn.Ready():
			return conn, nil
		case <-conn.Dead():
			c.dead(conn, nil)
			goto retry
		}
	}

	// 3rd case: no free connections, can't create yet one, wait for free.
	key, ch := c.freeReq.request()
	c.mu.Unlock()
	c.log.Debug(ctx, "Waiting for free connect", log.Int64("request_id", int64(key)))

	select {
	case conn := <-ch:
		c.log.Debug(ctx, "Got connection for request",
			log.Int64("conn_id", conn.id),
			log.Int64("request_id", int64(key)),
		)
		return conn, nil
	case <-c.stuck.Ready():
		c.log.Debug(ctx, "Some connection dead, try to create new connection, cancel waiting")

		c.freeReq.delete(key)
		select {
		default:
		case conn, ok := <-ch:
			if ok && conn != nil {
				return conn, nil
			}
		}

		goto retry
	case <-ctx.Done():
		err = ctx.Err()
	case <-c.ctx.Done():
		err = errors.Wrap(c.ctx.Err(), "DC closed")
	}

	// Executed only if at least one of context is Done.
	c.freeReq.delete(key)
	select {
	default:
	case conn, ok := <-ch:
		if ok && conn != nil {
			c.release(conn)
		}
	}

	return nil, err
}

// Invoke sends MTProto request using one of pool connection.
func (c *DC) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if c.closed.Load() {
		return errDCIsClosed
	}

	for {
		conn, err := c.acquire(ctx)
		if err != nil {
			if errors.Is(err, ErrConnDead) {
				continue
			}
			return errors.Wrap(err, "acquire connection")
		}

		c.log.Debug(ctx, "DC Invoke")
		err = conn.Invoke(ctx, input, output)
		if err != nil && errRetryableOnNewConn(err, c.retryOnWriteFailed) && ctx.Err() == nil {
			// Connection died before the request was processed by the
			// server, so it is safe to retry it on another connection.
			//
			// Mark connection as dead instead of releasing, so it is not
			// re-acquired and waiters are unblocked to create a new one.
			c.dead(conn, err)
			c.log.Debug(ctx, "DC Invoke failed on dead connection, retrying",
				log.Int64("conn_id", conn.id),
				log.Error(err),
			)
			continue
		}
		c.release(conn)
		if err != nil {
			c.log.Debug(ctx, "DC Invoke failed", log.Error(err))
			return errors.Wrap(err, "invoke pool")
		}

		c.log.Debug(ctx, "DC Invoke complete")
		return err
	}
}

// Close waits while all ongoing requests will be done or until given context is done.
// Then, closes the DC.
func (c *DC) Close() error {
	if c.closed.Swap(true) {
		return errors.New("DC already closed")
	}
	ctx := c.ctx
	c.log.Debug(ctx, "Closing DC")
	defer c.log.Debug(ctx, "DC closed")

	c.cancel()
	return c.grp.Wait()
}
