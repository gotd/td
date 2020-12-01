package telegram

import (
	"context"
	"crypto/rand"
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

type UpdateClient interface {
	RandInt64() (int64, error)
	SendMessage(ctx context.Context, m *tg.MessagesSendMessageRequest) error
}

// UpdateHandler will be called on received updates from Telegram.
type UpdateHandler func(ctx context.Context, c UpdateClient, u *tg.Updates) error

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg *tg.Client

	// conn is owned by Client and not exposed.
	// Currently immutable.
	conn net.Conn

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

func (c *Client) AuthKey() crypto.AuthKey {
	return c.authKey
}

// Options of Client.
type Options struct {
	// Required options:

	// PublicKeys of telegram.
	PublicKeys []*rsa.PublicKey
	// Addr to connect.
	Addr string

	// AppID is api_id of your application.
	///
	// Can be found on https://my.telegram.org/apps.
	AppID int
	// AppHash is api_hash of your application.
	//
	// Can be found on https://my.telegram.org/apps.
	AppHash string

	// Optional:

	// Dialer to use. Default dialer will be used if not provided.
	Dialer *net.Dialer
	// Network to use. Defaults to tcp.
	Network string
	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// SessionStorage will be used to load and save session data.
	// NB: Very sensitive data, save with care.
	SessionStorage SessionStorage
	// UpdateHandler will be called on received update.
	UpdateHandler UpdateHandler
}

// Dial initializes Client and creates connection to Telegram, initializing
// new or loading session from provided storage.
func Dial(ctx context.Context, opt Options) (*Client, error) {
	if opt.Dialer == nil {
		opt.Dialer = &net.Dialer{}
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	if opt.Random == nil {
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	if len(opt.PublicKeys) == 0 {
		// Using public keys that are included with distribution if not
		// provided.
		//
		// This should never fail and keys should be valid for recent
		// library versions.
		keys, err := vendoredKeys()
		if err != nil {
			return nil, xerrors.Errorf("failed to load vendored keys: %w", err)
		}
		opt.PublicKeys = keys
	}
	if opt.AppHash == "" {
		return nil, xerrors.New("no AppHash provided")
	}
	if opt.AppID == 0 {
		return nil, xerrors.New("no AppID provided")
	}

	conn, err := opt.Dialer.DialContext(ctx, "tcp", opt.Addr)
	if err != nil {
		return nil, xerrors.Errorf("failed to dial: %w", err)
	}
	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		conn:  conn,
		clock: time.Now,
		rand:  opt.Random,
		log:   opt.Logger,
		ping:  map[int64]func(){},
		rpc:   map[int64]func(b *bin.Buffer, rpcErr error){},

		ctx:    clientCtx,
		cancel: clientCancel,

		appID:   opt.AppID,
		appHash: opt.AppHash,

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

	// Loading session from storage if provided.
	if err := client.loadSession(ctx); err != nil {
		// TODO: Add opt-in config to ignore session load failures.
		return nil, xerrors.Errorf("failed to load session: %w", err)
	}

	// Starting connection.
	//
	// This will send initial packet to telegram and perform key exchange
	// if needed.
	if err := client.connect(ctx); err != nil {
		return nil, xerrors.Errorf("failed to start connection: %w", err)
	}

	if err := client.initConnection(ctx); err != nil {
		return nil, xerrors.Errorf("failed to init connection: %w", err)
	}

	return client, nil
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

	// Spawning reading goroutine.
	// Probably we should use another ctx here.
	go c.readLoop(c.ctx)

	return nil
}
