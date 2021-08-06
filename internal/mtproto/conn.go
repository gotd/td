package mtproto

import (
	"context"
	"crypto/rsa"
	"io"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mtproto/salts"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tdsync"
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

// MessageBuf is message id buffer.
type MessageBuf interface {
	Consume(id int64) bool
}

// Cipher handles message encryption and decryption.
type Cipher interface {
	DecryptFromBuffer(k crypto.AuthKey, buf *bin.Buffer) (*crypto.EncryptedMessageData, error)
	Encrypt(key crypto.AuthKey, data crypto.EncryptedMessageData, b *bin.Buffer) error
}

// Dialer is a abstraction for MTProto transport connection creator.
type Dialer func(ctx context.Context) (transport.Conn, error)

// Conn represents a MTProto client to Telegram.
type Conn struct {
	dialer        Dialer
	conn          transport.Conn
	handler       Handler
	rpc           *rpc.Engine
	rsaPublicKeys []*rsa.PublicKey
	types         *tmap.Map

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock        clock.Clock
	rand         io.Reader
	cipher       Cipher
	log          *zap.Logger
	messageID    MessageIDSource
	messageIDBuf MessageBuf // replay attack protection

	// use session() to access authKey, salt or sessionID.
	sessionMux sync.RWMutex
	authKey    crypto.AuthKey
	salt       int64
	sessionID  int64

	// server salts fetched by getSalts.
	salts salts.Salts

	// sentContentMessages is count of created content messages, used to
	// compute sequence number within session.
	sentContentMessages int32
	reqMux              sync.Mutex

	// ackSendChan is queue for outgoing message id's that require waiting for
	// ack from server.
	ackSendChan  chan int64
	ackBatchSize int
	ackInterval  time.Duration

	// callbacks for ping results.
	// Key is ping id.
	ping    map[int64]chan struct{}
	pingMux sync.Mutex
	// pingTimeout sets ping_delay_disconnect delay.
	pingTimeout time.Duration
	// pingInterval is duration between ping_delay_disconnect request.
	pingInterval time.Duration

	// gotSession is a signal channel for wait for handleSessionCreated message.
	gotSession *tdsync.Ready

	// handlers is waiter of all message handlers.
	handlers sync.WaitGroup

	// compressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using gzip.
	compressThreshold int
	dialTimeout       time.Duration
	exchangeTimeout   time.Duration
	saltFetchInterval time.Duration
	getTimeout        func(req uint32) time.Duration
	closed            atomic.Bool
}

// New creates new unstarted connection.
func New(dialer Dialer, opt Options) *Conn {
	// Set default values, if user does not set.
	opt.setDefaults()

	conn := &Conn{
		dialer:       dialer,
		clock:        opt.Clock,
		rand:         opt.Random,
		cipher:       opt.Cipher,
		log:          opt.Logger,
		messageID:    opt.MessageID,
		messageIDBuf: proto.NewMessageIDBuf(100),

		ackSendChan:  make(chan int64),
		ackInterval:  opt.AckInterval,
		ackBatchSize: opt.AckBatchSize,

		rsaPublicKeys: opt.PublicKeys,
		handler:       opt.Handler,
		types:         opt.Types,

		authKey: opt.Key,
		salt:    opt.Salt,

		ping:         map[int64]chan struct{}{},
		pingTimeout:  opt.PingTimeout,
		pingInterval: opt.PingInterval,

		gotSession: tdsync.NewReady(),

		rpc:               opt.engine,
		compressThreshold: 1024, // 1 KB
		dialTimeout:       opt.DialTimeout,
		exchangeTimeout:   opt.ExchangeTimeout,
		saltFetchInterval: opt.SaltFetchInterval,
		getTimeout:        opt.RequestTimeout,
	}
	if conn.rpc == nil {
		conn.rpc = rpc.New(conn.writeContentMessage, rpc.Options{
			Logger:        opt.Logger.Named("rpc"),
			RetryInterval: opt.RetryInterval,
			MaxRetries:    opt.MaxRetries,
			Clock:         opt.Clock,
			DropHandler:   conn.dropRPC,
		})
	}

	return conn
}

// handleClose closes rpc engine and underlying connection on context done.
func (c *Conn) handleClose(ctx context.Context) error {
	<-ctx.Done()
	c.log.Debug("Closing")

	// Close RPC Engine.
	c.rpc.ForceClose()
	// Close connection.
	if err := c.conn.Close(); err != nil {
		c.log.Debug("Failed to cleanup connection", zap.Error(err))
	}
	c.closed.Store(true)
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
	if c.closed.Load() {
		return xerrors.New("do Run on closed connection")
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
		g := tdsync.NewLogGroup(ctx, c.log.Named("group"))
		g.Go("handleClose", c.handleClose)
		g.Go("pingLoop", c.pingLoop)
		g.Go("ackLoop", c.ackLoop)
		g.Go("saltsLoop", c.saltLoop)
		g.Go("userCallback", f)
		g.Go("readLoop", c.readLoop)
		g.Go("handlers", func(context.Context) error {
			// Wait all spawned handlers.
			c.handlers.Wait()
			return nil
		})

		if err := g.Wait(); err != nil {
			return xerrors.Errorf("group: %w", err)
		}
	}
	return nil
}

// connect establishes connection using configured transport, creating
// new auth key if needed.
func (c *Conn) connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()

	conn, err := c.dialer(ctx)
	if err != nil {
		return xerrors.Errorf("dial failed: %w", err)
	}
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
