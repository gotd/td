package telegram

import (
	"context"
	"crypto/rsa"
	"io"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
)

// UpdateHandler will be called on received updates from Telegram.
type UpdateHandler func(ctx context.Context, u *tg.Updates) error

// Available MTProto default server addresses.
//
// See https://my.telegram.org/apps.
const (
	AddrProduction = "149.154.167.50:443"
	AddrTest       = "149.154.167.40:443"
)

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg *tg.Client

	// conn is owned by Client and not exposed.
	conn   net.Conn
	addr   string
	dialer Dialer

	pingDuration time.Duration
	pingTimeout  time.Duration

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock func() time.Time
	rand  io.Reader
	log   *zap.Logger

	// Access to authKey and authKeyID is not synchronized because
	// serial access ensured in Dial (i.e. no concurrent access possible).
	authKey   crypto.AuthKey
	authKeyID [8]byte

	salt    int64 // atomic access only
	session int64 // atomic access only

	// sentContentMessages is count of created content messages, used to
	// compute sequence number within session.
	//
	// protected by sentContentMessagesMux.
	sentContentMessages    int32
	sentContentMessagesMux sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	appID   int    // immutable
	appHash string // immutable

	updateHandler  UpdateHandler  // immutable
	sessionStorage SessionStorage // immutable

	// callbacks for RPC requests, protected by rpcMux
	rpc    map[int64]func(b *bin.Buffer, rpcErr error)
	rpcMux sync.Mutex

	// callbacks for ping results protected by pingMux
	ping    map[int64]func()
	pingMux sync.Mutex

	// immutable
	rsaPublicKeys []*rsa.PublicKey

	types *tmap.Map
}

const defaultTimeout = time.Second * 10

func (c *Client) startIntermediateMode(deadline time.Time) error {
	if err := c.conn.SetDeadline(deadline); err != nil {
		return xerrors.Errorf("failed to set deadline: %w", err)
	}
	if _, err := c.conn.Write(proto.IntermediateClientStart); err != nil {
		return xerrors.Errorf("failed to write start: %w", err)
	}
	if err := c.conn.SetDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("failed to reset connection deadline: %w", err)
	}
	return nil
}

func (c *Client) resetDeadline() error {
	return c.conn.SetDeadline(time.Time{})
}

func (c *Client) deadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}
	return c.clock().Add(defaultTimeout)
}

func (c *Client) newUnencryptedMessage(payload bin.Encoder, b *bin.Buffer) error {
	b.Reset()
	if err := payload.Encode(b); err != nil {
		return err
	}
	msg := proto.UnencryptedMessage{
		MessageID:   c.newMessageID(),
		MessageData: b.Copy(),
	}
	b.Reset()
	return msg.Encode(b)
}

// NewClient creates new unstarted client.
func NewClient(appID int, appHash string, opt Options) *Client {
	// Set default values, if user does not set.
	opt.setDefaults()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		addr:   opt.Addr,
		dialer: opt.Dialer,

		pingDuration: opt.PingDuration,
		pingTimeout:  opt.PingTimeout,

		clock: time.Now,
		rand:  opt.Random,
		log:   opt.Logger,
		ping:  map[int64]func(){},
		rpc:   map[int64]func(b *bin.Buffer, rpcErr error){},

		ctx:    clientCtx,
		cancel: clientCancel,

		appID:   appID,
		appHash: appHash,

		sessionStorage: opt.SessionStorage,
		rsaPublicKeys:  opt.PublicKeys,
		updateHandler:  opt.UpdateHandler,

		types: tmap.New(
			mt.TypesMap(),
			tg.TypesMap(),
			proto.TypesMap(),
		),
	}

	// Initializing internal RPC caller.
	client.tg = tg.NewClient(client)

	return client
}

// Connect initializes connection to Telegram server and starts internal
// read loop.
func (c *Client) Connect(ctx context.Context) (err error) {
	c.conn, err = c.dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return xerrors.Errorf("failed to dial: %w", err)
	}

	// Loading session from storage if provided.
	if err := c.loadSession(ctx); err != nil {
		// TODO: Add opt-in config to ignore session load failures.
		return xerrors.Errorf("failed to load session: %w", err)
	}

	// Starting connection.
	//
	// This will send initial packet to telegram and perform key exchange
	// if needed.
	if err := c.connect(ctx); err != nil {
		return xerrors.Errorf("failed to start connection: %w", err)
	}

	// Spawning reading goroutine.
	go c.readLoop(c.ctx)

	// Spawning ping goroutine.
	go c.pingLoop(ctx)

	if err := c.initConnection(ctx); err != nil {
		return xerrors.Errorf("failed to init connection: %w", err)
	}

	return nil
}

// Authenticated returns true of already authenticated.
func (c *Client) Authenticated() bool {
	return !c.authKey.Zero()
}

// connect establishes connection in intermediate mode, creating new auth key
// if needed.
func (c *Client) connect(ctx context.Context) error {
	deadline := c.clock().Add(defaultTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}
	if err := c.startIntermediateMode(deadline); err != nil {
		return xerrors.Errorf("failed to initialize intermediate protocol: %w", err)
	}
	if c.authKey.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock()
		if err := c.createAuthKey(ctx); err != nil {
			return xerrors.Errorf("unable to create auth key: %w", err)
		}
		c.log.With(zap.Duration("duration", c.clock().Sub(start))).Info("Auth key generated")
	}
	return nil
}
