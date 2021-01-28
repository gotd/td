package pool

import (
	"context"
	"sync"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

// DCOptions is a Telegram data center connections pool options.
type DCOptions struct {
	// InitConnection parameters.
	// AppID of Telegram application.
	AppID int
	// Telegram device information.
	Device DeviceConfig

	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// MTProto options for connections.
	MTProto mtproto.Options
	// Opened connection limit to the DC.
	MaxOpenConnections int64
}

// DC represents connection pool to one data center.
type DC struct {
	id dcID
	// MTProto connection options.
	addr string          // immutable
	opts mtproto.Options // immutable
	// MTProto session parameters. Unique for DC.
	authKey    crypto.AuthKey
	salt       int64
	sessionMux sync.Mutex

	// InitConnection parameters.
	appID int // immutable
	// Telegram device information.
	device DeviceConfig // immutable

	// Wrappers for external world, like logs or PRNG.
	log *zap.Logger // immutable

	// Handler passed by client.
	handler ConnHandler // immutable

	// DC context. Will be canceled by Run on exit.
	ctx    context.Context    // immutable
	cancel context.CancelFunc // immutable

	// Connections supervisor.
	grp *tdsync.Supervisor
	// Free connections.
	free    []*poolConn
	freeMux sync.Mutex
	freeReq *reqMap

	// Total connections.
	total atomic.Int64
	// Limit of connections.
	max int64 // immutable

	// Requests wait group.
	ongoing sync.WaitGroup

	ready       *tdsync.Ready
	closed      atomic.Bool
	nextRequest atomic.Int64
}

// NewDC creates new uninitialized DC.
func NewDC(id dcID, addr string, handler ConnHandler, opts DCOptions) *DC {
	ctx, cancel := context.WithCancel(context.Background())

	return &DC{
		id:      id,
		addr:    addr,
		opts:    opts.MTProto,
		appID:   opts.AppID,
		device:  opts.Device,
		handler: handler,
		log:     opts.Logger,
		ctx:     ctx,
		cancel:  cancel,
		grp:     tdsync.NewSupervisor(ctx),
		freeReq: newReqMap(),
		max:     opts.MaxOpenConnections,
		ready:   tdsync.NewReady(),
	}
}

// OnSession implements ConnHandler.
func (c *DC) OnSession(addr string, cfg tg.Config, s mtproto.Session) error {
	c.log.Debug("Session created")

	c.sessionMux.Lock()
	c.salt = s.Salt
	c.authKey = s.Key
	c.sessionMux.Unlock()
	c.ready.Signal()

	return c.handler.OnSession(addr, cfg, s)
}

// OnMessage implements ConnHandler.
func (c *DC) OnMessage(b *bin.Buffer) error {
	return c.handler.OnMessage(b)
}

func (c *DC) createConnection(mode connMode) *poolConn {
	c.log.Debug(
		"Creating new connection",
		zap.String("addr", c.addr),
		zap.Int64("total", c.total.Load()),
	)

	opts := c.opts
	opts.Logger = c.log.Named("conn").With(zap.Int64("conn_id", c.nextRequest.Inc()))
	conn := &poolConn{
		conn: newConn(c, c.addr, mode, opts),
		dc:   c,
	}

	c.grp.Go(func(groupCtx context.Context) error {
		c.total.Inc()
		defer c.total.Dec()
		defer c.dead(conn)

		return conn.Init(groupCtx, c.appID, c.device)
	})

	return conn
}

func (c *DC) dead(r *poolConn) {
	c.log.Debug(
		"Dead connection",
		zap.String("addr", c.addr),
		zap.Int64("total", c.total.Load()),
	)

	r.dead.Store(true)
	c.freeMux.Lock()
	defer c.freeMux.Unlock()

	idx := -1
	for i, conn := range c.free {
		// Search connection by pointer.
		if conn == r {
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
}

func (c *DC) pop() (r *poolConn, ok bool) {
	c.freeMux.Lock()
	defer c.freeMux.Unlock()

	l := len(c.free)
	if l > 0 {
		r, c.free = c.free[l-1], c.free[:l-1]

		return r, true
	}

	return
}

func (c *DC) release(r *poolConn) {
	if c.freeReq.transfer(r) {
		return
	}
	c.freeMux.Lock()
	c.free = append(c.free, r)
	c.freeMux.Unlock()
}

var errDCIsClosed = xerrors.New("DC is closed")

func (c *DC) acquire(ctx context.Context) (r *poolConn, err error) {
	// 1st case: have free connections.
	if r, ok := c.pop(); ok {
		return r, nil
	}

	// 2nd case: no free connections, but can create one.
	// c.max < 1 means unlimited
	if c.max < 1 || c.total.Load() < c.max {
		conn := c.createConnection(connModeUpdates)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.ctx.Done():
			return nil, xerrors.Errorf("DC closed: %w", c.ctx.Err())
		case <-conn.Ready():
			return conn, nil
		}
	}

	// 3rd case: no free connections, can't create yet one, wait for free.
	key, ch := c.freeReq.request()

	select {
	case conn := <-ch:
		return conn, nil

	case <-ctx.Done():
		err = ctx.Err()
	case <-c.ctx.Done():
		err = xerrors.Errorf("DC closed: %w", c.ctx.Err())
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

// Ready sends ready signal when DC is initialized.
func (c *DC) Ready() <-chan struct{} {
	return c.ready.Ready()
}

// Run initialize connection pool.
func (c *DC) Run(ctx context.Context) error {
	if c.closed.Load() {
		return errDCIsClosed
	}

	conn, err := c.acquire(ctx)
	if err != nil {
		return err
	}
	c.release(conn)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return nil
	}
}

// InvokeRaw sends MTProto request using one of pool connection.
func (c *DC) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if c.closed.Load() {
		return errDCIsClosed
	}

	c.ongoing.Add(1)
	defer c.ongoing.Done()

	for {
		conn, err := c.acquire(ctx)
		if err != nil {
			return xerrors.Errorf("acquire connection: %w", err)
		}

		err = conn.InvokeRaw(ctx, input, output)
		if conn.dead.Load() {
			continue
		}

		c.release(conn)
		return err
	}
}

// Close waits while all ongoing requests will be done or until given context is done.
// Then, closes the pool.
func (c *DC) Close(closeCtx context.Context) error {
	if c.closed.Swap(true) {
		return xerrors.New("DC already closed")
	}
	c.log.Info("DC connection pool closing")

	closed, cancel := context.WithCancel(closeCtx)
	go func() {
		c.ongoing.Wait()
		cancel()
	}()

	<-closed.Done()

	c.cancel()
	return c.grp.Wait()
}
