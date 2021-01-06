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
	Key  crypto.AuthKey
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

// newSessionID sets session id to random value.
func (c *Conn) newSessionID() error {
	id, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	c.sessionMux.Lock()
	defer c.sessionMux.Unlock()
	c.sessionID = id

	return nil
}

// Cipher handles message encryption and decryption.
type Cipher interface {
	DecryptFromBuffer(k crypto.AuthKey, buf *bin.Buffer) (*crypto.EncryptedMessageData, error)
	Encrypt(key crypto.AuthKey, data crypto.EncryptedMessageData, b *bin.Buffer) error
}

// Conn represents a MTProto client to Telegram.
type Conn struct {
	transport     Transport
	conn          transport.Conn
	addr          string
	handler       Handler
	rpc           *rpc.Engine
	rsaPublicKeys []*rsa.PublicKey
	types         *tmap.Map

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock     clock.Clock
	rand      io.Reader
	cipher    Cipher
	log       *zap.Logger
	messageID MessageIDSource

	// use session() to access authKey, salt or sessionID.
	sessionMux sync.RWMutex
	authKey    crypto.AuthKey
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

	readConcurrency int
	messages        chan *crypto.EncryptedMessageData

	closed bool
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
		cipher:    opt.Cipher,
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

		readConcurrency: opt.ReadConcurrency,
		messages:        make(chan *crypto.EncryptedMessageData, opt.ReadConcurrency),

		rpc: opt.engine,
	}
	if conn.rpc == nil {
		conn.rpc = rpc.New(conn.write, rpc.Options{
			Logger:        opt.Logger.Named("rpc"),
			RetryInterval: opt.RetryInterval,
			MaxRetries:    opt.MaxRetries,
			Clock:         opt.Clock,
		})
	}

	return conn
}

func goGroup(ctx context.Context, g *errgroup.Group, f func(ctx context.Context) error) {
	g.Go(func() error {
		return f(ctx)
	})
}

// handleClose closes rpc engine and underlying connection on context done.
func (c *Conn) handleClose(ctx context.Context) error {
	<-ctx.Done()
	c.log.Debug("Closing")
	c.rpc.Close()
	if err := c.conn.Close(); err != nil {
		c.log.Debug("Failed to cleanup connection", zap.Error(err))
	}
	c.closed = true
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
	if c.closed {
		return xerrors.New("failed to Run closed connection")
	}
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
		for i := 0; i < c.readConcurrency; i++ {
			g.Go(c.readEncryptedMessages)
		}
		if err := g.Wait(); err != nil {
			return xerrors.Errorf("group: %w", err)
		}
	}
	return nil
}

// connect establishes connection using configured transport, creating
// new auth key if needed.
func (c *Conn) connect(ctx context.Context) error {
	conn, err := c.transport.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return xerrors.Errorf("dial failed: %w", err)
	}
	c.log.Info("Dialed transport", zap.String("addr", c.addr))
	c.conn = conn

	session := c.session()

	if session.Key.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("create auth key: %w", err)
		}

		c.log.Info("Auth key generated",
			zap.Duration("duration", c.clock.Now().Sub(start)),
		)
		return nil
	}

	c.log.Info("Key already exists")
	if session.ID == 0 {
		// NB: Telegram can return 404 error if session id is zero.
		//
		// See https://github.com/gotd/td/issues/107.
		c.log.Debug("Generating new session id")
		if err := c.newSessionID(); err != nil {
			return err
		}
	}

	return nil
}
