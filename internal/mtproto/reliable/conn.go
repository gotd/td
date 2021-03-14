package reliable

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/lifetime"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/tdsync"
)

// Conn is a reliable MTProto connection.
type Conn struct {
	addr        string
	opts        mtproto.Options
	createConn  func(addr string, opts mtproto.Options) MTConn
	onConnected func(MTConn) error

	conn MTConn

	mux sync.RWMutex
	log *zap.Logger
}

// New creates new reliable MTProto conn.
func New(cfg Config) *Conn {
	cfg.setDefaults()
	opts := cfg.MTOpts

	log := opts.Logger
	if log == nil {
		log = zap.NewNop()
	}

	if opts.SessionHandler == nil {
		opts.SessionHandler = func(session mtproto.Session) error { return nil }
	}

	conn := &Conn{
		addr:        cfg.Addr,
		opts:        opts,
		createConn:  cfg.CreateConn,
		onConnected: cfg.OnConnected,
		log:         log.Named("reli"),
	}

	conn.opts.SessionHandler = conn.wrapSessionHandler(conn.opts.SessionHandler)
	return conn
}

// Run starts the connection.
func (c *Conn) Run(ctx context.Context, f func(context.Context) error) error {
	life, err := c.connect(1)
	if err != nil {
		return err
	}

	g := tdsync.NewCancellableGroup(ctx)
	defer g.Cancel()

	g.Go(func(ctx context.Context) error {
		return f(ctx)
	})

	g.Go(func(ctx context.Context) error {
		return c.loop(ctx, life, 5)
	})

	return g.Wait()
}

// InvokeRaw sens input and decodes result into output.
func (c *Conn) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	c.mux.RLock()
	conn := c.conn
	c.mux.RUnlock()

	// TODO(ccln): Check request status.
	return conn.InvokeRaw(ctx, in, out)
}

func (c *Conn) loop(ctx context.Context, life *lifetime.Life, maxAttempts int) error {
waitUntilDisconnect:

	e := make(chan error)
	go func() { e <- life.Wait() }()

	select {
	case err := <-e:
		if err == nil {
			c.log.Info("Disconnected")
			return nil
		}

		c.log.Warn("Connection error", zap.Error(err))
	case <-ctx.Done():
		c.log.Info("Forced exit", zap.Error(life.Stop()))
		return ctx.Err()
	}

	c.log.Info("Reconnecting")
	var err error
	life, err = c.connect(maxAttempts)
	if err != nil {
		return err
	}

	goto waitUntilDisconnect
}

func (c *Conn) connect(maxAttempts int) (*lifetime.Life, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// TODO(ccln): Backoff.
	attempt := 0
retry:
	conn := c.createConn(c.addr, c.opts)
	life, err := lifetime.Start(conn)
	if err != nil {
		c.log.Warn("Failed to connect to the server", zap.Error(err), zap.Int("attempt", attempt))
		if attempt == maxAttempts {
			return nil, err
		}

		time.Sleep(time.Second)
		attempt++
		goto retry
	}

	if err := c.onConnected(conn); err != nil {
		return nil, err
	}

	c.conn = conn
	return life, nil
}

// Keep credentials up-to-date between reconnections.
func (c *Conn) wrapSessionHandler(f func(mtproto.Session) error) func(mtproto.Session) error {
	return func(s mtproto.Session) error {
		c.mux.Lock()
		// TODO(ccln): session id?
		c.opts.Key = s.Key
		c.opts.Salt = s.Salt
		c.mux.Unlock()
		return f(s)
	}
}
