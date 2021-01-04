package mtproto

import (
	"context"
	"crypto/rsa"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/transport"
)

// Handler will be called on received message from Telegram.
type Handler func(b *bin.Buffer) error

// MessageIDSource is message id generator.
type MessageIDSource interface {
	New(t proto.MessageType) int64
}

// Session represents connection state.
type Session struct {
	Key  crypto.AuthKeyWithID
	Salt int64
}

// Session returns current connection session info.
func (c *Conn) Session() Session {
	return Session{
		Key:  c.authKey,
		Salt: c.salt,
	}
}

// Conn represents a MTProto client to Telegram.
type Conn struct {
	transport   Transport
	conn        transport.Conn
	addr        string
	trace       tracer
	onReconnect func() error

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock     clock.Clock
	rand      io.Reader
	cipher    crypto.Cipher
	log       *zap.Logger
	messageID MessageIDSource

	authKey   crypto.AuthKeyWithID
	salt      int64 // atomic access only
	sessionID int64 // atomic access only

	// sentContentMessages is count of created content messages, used to
	// compute sequence number within session.
	sentContentMessages    int32
	sentContentMessagesMux sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	handler Handler // immutable

	rpc *rpc.Engine

	// ackSendChan is queue for outgoing message id's that require waiting for
	// ack from server.
	ackSendChan  chan int64
	ackBatchSize int
	ackInterval  time.Duration

	// callbacks for ping results.
	// Key is ping id.
	ping    map[int64]func()
	pingMux sync.Mutex

	// immutable
	rsaPublicKeys []*rsa.PublicKey

	types *tmap.Map
}

// NewConn creates new unstarted connection.
func NewConn(addr string, opt Options) *Conn {
	// Set default values, if user does not set.
	opt.setDefaults()

	connCtx, connCancel := context.WithCancel(context.Background())
	conn := &Conn{
		addr:        addr,
		transport:   opt.Transport,
		onReconnect: opt.OnReconnect,
		clock:       opt.Clock,
		rand:        opt.Random,
		cipher:      crypto.NewClientCipher(opt.Random),
		log:         opt.Logger,
		ping:        map[int64]func(){},
		messageID:   opt.MessageID,

		ackSendChan:  make(chan int64),
		ackInterval:  opt.AckInterval,
		ackBatchSize: opt.AckBatchSize,

		ctx:    connCtx,
		cancel: connCancel,

		rsaPublicKeys: opt.PublicKeys,
		handler:       opt.Handler,
		types:         opt.Types,

		authKey: opt.Key,
		salt:    opt.Salt,
	}
	conn.rpc = rpc.New(conn.write, rpc.Config{
		Logger:        opt.Logger.Named("rpc"),
		RetryInterval: opt.RetryInterval,
		MaxRetries:    opt.MaxRetries,
	})

	return conn
}

// Connect initializes connection to Telegram server and starts internal
// read loop.
func (c *Conn) Connect(ctx context.Context) (err error) {
	// Starting connection.
	//
	// This will send initial packet to telegram and perform key exchange
	// if needed.
	if err := c.connect(ctx); err != nil {
		return xerrors.Errorf("start: %w", err)
	}

	// Spawning goroutines.
	go c.readLoop(c.ctx)
	go c.ackLoop(c.ctx)
	go c.pingLoop(c.ctx)

	return nil
}

// Reconnect performs re-connection. Same as Connect, but does not
// start new goroutines.
func (c *Conn) Reconnect(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
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
	c.log.Info("Connected", zap.String("addr", c.addr))
	c.conn = conn

	if c.authKey.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("create auth key: %w", err)
		}

		c.log.With(
			zap.Duration("duration", c.clock.Now().Sub(start)),
		).Info("Auth key generated")
	}
	return nil
}
