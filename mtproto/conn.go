package mtproto

import (
	"context"
	"crypto/rsa"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/transport"
)

// Handler will be called on received message from Telegram.
type Handler interface {
	OnMessage(b *bin.Buffer) error
	OnSession(session Session) error
}

// MessageIDSource is message id generator.
type MessageIDSource interface {
	New(t proto.MessageType) int64
}

// Session represents connection state.
type Session struct {
	ID   int64
	Key  crypto.AuthKeyWithID
	Salt int64
}

// Session returns current connection session info.
func (c *Conn) session() Session {
	c.sessionMux.RLock()
	defer c.sessionMux.RUnlock()
	return Session{
		Key:  c.authKey,
		Salt: c.salt,
		ID:   c.sessionID,
	}
}

// Conn represents a MTProto client to Telegram.
type Conn struct {
	transport     Transport
	conn          transport.Conn
	addr          string
	trace         tracer
	handler       Handler
	rpc           *rpc.Engine
	rsaPublicKeys []*rsa.PublicKey
	types         *tmap.Map

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock     clock.Clock
	rand      io.Reader
	cipher    crypto.Cipher
	log       *zap.Logger
	messageID MessageIDSource

	// use session() to access authKey, salt or sessionID.
	sessionMux sync.RWMutex
	authKey    crypto.AuthKeyWithID
	salt       int64
	sessionID  int64

	// sentContentMessages is count of created content messages, used to
	// compute sequence number within session.
	sentContentMessages    int32
	sentContentMessagesMux sync.Mutex

	// ackSendChan is queue for outgoing message id's that require waiting for
	// ack from server.
	ackSendChan  chan int64
	ackBatchSize int
	ackInterval  time.Duration

	// callbacks for ping results.
	// Key is ping id.
	ping    map[int64]func()
	pingMux sync.Mutex
}

// New creates new unstarted connection.
func New(addr string, opt Options) *Conn {
	// Set default values, if user does not set.
	opt.setDefaults()

	conn := &Conn{
		addr:      addr,
		transport: opt.Transport,
		clock:     opt.Clock,
		rand:      opt.Random,
		cipher:    crypto.NewClientCipher(opt.Random),
		log:       opt.Logger,
		ping:      map[int64]func(){},
		messageID: opt.MessageID,

		ackSendChan:  make(chan int64),
		ackInterval:  opt.AckInterval,
		ackBatchSize: opt.AckBatchSize,

		rsaPublicKeys: opt.PublicKeys,
		handler:       opt.Handler,
		types:         opt.Types,

		authKey: opt.Key,
		salt:    opt.Salt,
	}
	conn.rpc = rpc.New(conn.write, rpc.Options{
		Logger:        opt.Logger.Named("rpc"),
		RetryInterval: opt.RetryInterval,
		MaxRetries:    opt.MaxRetries,
		Clock:         opt.Clock,
	})

	return conn
}

func goGroup(ctx context.Context, g *errgroup.Group, f func(ctx context.Context) error) {
	g.Go(func() error {
		return f(ctx)
	})
}

func (c *Conn) handleClose(ctx context.Context) error {
	<-ctx.Done()
	c.rpc.Close()
	if err := c.conn.Close(); err != nil {
		c.log.Debug("Failed to cleanup connection", zap.Error(err))
	}
	return nil
}

// Run initializes MTProto connection to server and blocks until disconnection.
//
// When connection is ready, Handler.OnSession is called.
func (c *Conn) Run(ctx context.Context, f func(ctx context.Context) error) error {
	// Starting connection.
	//
	// This will send initial packet to telegram and perform key exchange
	// if needed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.log.Debug("Run: start")
	defer c.log.Debug("Run: end")
	if err := c.connect(ctx); err != nil {
		return xerrors.Errorf("start: %w", err)
	}
	{
		// All goroutines are bound to current call.
		g, gCtx := errgroup.WithContext(ctx)
		goGroup(gCtx, g, c.handleClose)
		goGroup(gCtx, g, c.pingLoop)
		goGroup(gCtx, g, c.readLoop)
		goGroup(gCtx, g, c.ackLoop)
		goGroup(gCtx, g, f)
		if err := g.Wait(); err != nil {
			return xerrors.Errorf("group: %w", err)
		}
	}
	return nil
}

// connect establishes connection in intermediate mode, creating new auth key
// if needed.
func (c *Conn) connect(ctx context.Context) error {
	conn, err := c.transport.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return xerrors.Errorf("dial failed: %w", err)
	}
	c.log.Info("Dialed transport", zap.String("addr", c.addr))
	c.conn = conn

	if c.session().Key.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("create auth key: %w", err)
		}

		c.log.With(
			zap.Duration("duration", c.clock.Now().Sub(start)),
		).Info("Auth key generated")
	} else {
		c.log.Info("Key already exists")
	}
	return nil
}
