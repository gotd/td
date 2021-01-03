package mtproto

import (
	"context"
	"crypto/rsa"
	"io"
	"sync"
	"time"

	"github.com/gotd/td/internal/clock"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// Handler will be called on received message from Telegram.
type Handler func(b *bin.Buffer) error

// Available MTProto default server addresses.
//
// See https://my.telegram.org/apps.
const (
	AddrProduction = "149.154.167.50:443"
	AddrTest       = "149.154.167.40:443"
)

// Test-only credentials. Can be used with AddrTest and TestAuth to
// test authentication.
//
// Reference:
//	* https://github.com/telegramdesktop/tdesktop/blob/5f665b8ecb48802cd13cfb48ec834b946459274a/docs/api_credentials.md
const (
	TestAppID   = 17349
	TestAppHash = "344583e45741c457fe1862106095a5eb"
)

// MessageIDSource is message id generator.
type MessageIDSource interface {
	New(t proto.MessageType) int64
}

// Conn represents a MTProto client to Telegram.
type Conn struct {
	mux sync.Mutex

	// tg provides RPC calls via Conn.
	tg *tg.Client

	// conn is owned by Conn and not exposed.
	transport Transport
	connMux   sync.RWMutex
	conn      transport.Conn
	addr      string
	trace     tracer

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock     clock.Clock
	rand      io.Reader
	cipher    crypto.Cipher
	log       *zap.Logger
	messageID MessageIDSource

	sessionCreated *condOnce

	// Access to authKey and authKeyID is not synchronized because
	// serial access ensured in Dial (i.e. no concurrent access possible).
	authKey crypto.AuthKeyWithID

	salt    int64 // atomic access only
	session int64 // atomic access only

	// sentContentMessages is count of created content messages, used to
	// compute sequence number within session.
	sentContentMessages    int32
	sentContentMessagesMux sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	appID   int    // immutable
	appHash string // immutable

	handler        Handler        // immutable
	sessionStorage SessionStorage // immutable

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
func NewConn(appID int, appHash string, opt Options) *Conn {
	// Set default values, if user does not set.
	opt.setDefaults()

	now := time.Now

	const defaultMsgIDGenBuf = 100

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Conn{
		addr:      opt.Addr,
		transport: opt.Transport,

		clock:     opt.Clock,
		rand:      opt.Random,
		cipher:    crypto.NewClientCipher(opt.Random),
		log:       opt.Logger,
		ping:      map[int64]func(){},
		messageID: proto.NewMessageIDGen(now, defaultMsgIDGenBuf),

		sessionCreated: createCondOnce(),

		ackSendChan:  make(chan int64),
		ackInterval:  opt.AckInterval,
		ackBatchSize: opt.AckBatchSize,

		ctx:    clientCtx,
		cancel: clientCancel,

		appID:   appID,
		appHash: appHash,

		sessionStorage: opt.SessionStorage,
		rsaPublicKeys:  opt.PublicKeys,
		handler:        opt.Handler,

		types: tmap.New(
			mt.TypesMap(),
			tg.TypesMap(),
			proto.TypesMap(),
		),
	}

	client.rpc = rpc.New(client.write, rpc.Config{
		Logger:        opt.Logger.Named("rpc"),
		RetryInterval: opt.RetryInterval,
		MaxRetries:    opt.MaxRetries,
	})

	// Initializing internal RPC caller.
	client.tg = tg.NewClient(client)

	return client
}

// Connect initializes connection to Telegram server and starts internal
// read loop.
func (c *Conn) Connect(ctx context.Context) (err error) {
	// Loading session from storage if provided.
	if err := c.loadSession(ctx); err != nil {
		// TODO: Add opt-in config to ignore session load failures.
		return xerrors.Errorf("load session: %w", err)
	}

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

	if err := c.initConnection(ctx, connDefault); err != nil {
		return xerrors.Errorf("init: %w", err)
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

	c.mux.Lock()
	defer c.mux.Unlock()
	c.conn = conn

	if c.authKey.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("create auth key: %w", err)
		}

		if err := c.saveSession(ctx); err != nil {
			return xerrors.Errorf("failed to save session: %w", err)
		}

		c.log.With(zap.Duration("duration", c.clock.Now().Sub(start))).Info("AuthFlow key generated")
	}
	return nil
}
