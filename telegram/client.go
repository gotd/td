package telegram

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/oteltg"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/telegram/internal/version"
	"github.com/gotd/td/tg"
)

// UpdateHandler will be called on received updates from Telegram.
type UpdateHandler interface {
	Handle(ctx context.Context, u tg.UpdatesClass) error
}

// UpdateHandlerFunc type is an adapter to allow the use of
// ordinary function as update handler.
//
// UpdateHandlerFunc(f) is an UpdateHandler that calls f.
type UpdateHandlerFunc func(ctx context.Context, u tg.UpdatesClass) error

// Handle calls f(ctx, u)
func (f UpdateHandlerFunc) Handle(ctx context.Context, u tg.UpdatesClass) error {
	return f(ctx, u)
}

type clientStorage interface {
	Load(ctx context.Context) (*session.Data, error)
	Save(ctx context.Context, data *session.Data) error
}

type clientConn interface {
	Run(ctx context.Context) error
	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Ping(ctx context.Context) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	// Put migration in the header of the structure to ensure 64-bit alignment,
	// otherwise it will cause the atomic operation of connsCounter to panic.
	// DO NOT change the order of members arbitrarily.
	// Ref: https://pkg.go.dev/sync/atomic#pkg-note-BUG

	// Connection factory fields.
	connsCounter   atomic.Int64
	create         connConstructor        // immutable
	resolver       dcs.Resolver           // immutable
	onDead         func()                 // immutable
	newConnBackoff func() backoff.BackOff // immutable
	defaultMode    manager.ConnMode       // immutable

	// Migration state.
	migrationTimeout time.Duration // immutable
	migration        chan struct{}

	// tg provides RPC calls via Client. Uses invoker below.
	tg *tg.Client // immutable
	// invoker implements tg.Invoker on top of Client and mw.
	invoker tg.Invoker // immutable
	// mw is list of middlewares used in invoker, can be blank.
	mw []Middleware // immutable

	// Telegram device information.
	device DeviceConfig // immutable

	// MTProto options.
	opts mtproto.Options // immutable

	// DCList state.
	// Domain list (for websocket)
	domains map[int]string // immutable
	// Denotes to use Test DCs.
	testDC bool // immutable

	// Connection state. Guarded by connMux.
	session     *pool.SyncSession
	cfg         *manager.AtomicConfig
	conn        clientConn
	connBackoff atomic.Pointer[backoff.BackOff]
	connMux     sync.Mutex

	// Restart signal channel.
	restart chan struct{} // immutable

	// Connections to non-primary DC.
	subConns    map[int]CloseInvoker
	subConnsMux sync.Mutex
	sessions    map[int]*pool.SyncSession
	sessionsMux sync.Mutex

	// Wrappers for external world, like logs or PRNG.
	rand  io.Reader   // immutable
	log   *zap.Logger // immutable
	clock clock.Clock // immutable

	// Client context. Will be canceled by Run on exit.
	ctx    context.Context
	cancel context.CancelFunc

	// Client config.
	appID   int    // immutable
	appHash string // immutable
	// Session storage.
	storage clientStorage // immutable, nillable

	// Ready signal channel, sends signal when client connection is ready.
	// Resets on reconnect.
	ready *tdsync.ResetReady // immutable

	// Telegram updates handler.
	updateHandler UpdateHandler // immutable
	// Denotes that no update mode is enabled.
	noUpdatesMode bool // immutable

	// Tracing.
	tracer trace.Tracer

	// onTransfer is called in transfer.
	onTransfer AuthTransferHandler

	// onSelfError is called on error calling Self().
	onSelfError func(ctx context.Context, err error) error
}

// NewClient creates new unstarted client.
func NewClient(appID int, appHash string, opt Options) *Client {
	opt.setDefaults()

	mode := manager.ConnModeUpdates
	if opt.NoUpdates {
		mode = manager.ConnModeData
	}
	client := &Client{
		rand:          opt.Random,
		log:           opt.Logger,
		appID:         appID,
		appHash:       appHash,
		updateHandler: opt.UpdateHandler,
		session: pool.NewSyncSession(pool.Session{
			DC: opt.DC,
		}),
		domains: opt.DCList.Domains,
		testDC:  opt.DCList.Test,
		cfg: manager.NewAtomicConfig(tg.Config{
			DCOptions: opt.DCList.Options,
		}),
		create:           defaultConstructor(),
		resolver:         opt.Resolver,
		defaultMode:      mode,
		newConnBackoff:   opt.ReconnectionBackoff,
		onDead:           opt.OnDead,
		clock:            opt.Clock,
		device:           opt.Device,
		migrationTimeout: opt.MigrationTimeout,
		noUpdatesMode:    opt.NoUpdates,
		mw:               opt.Middlewares,
		onTransfer:       opt.OnTransfer,
		onSelfError:      opt.OnSelfError,
	}
	if opt.TracerProvider != nil {
		client.tracer = opt.TracerProvider.Tracer(oteltg.Name)
	}
	client.init()

	// Including version into client logger to help with debugging.
	if v := version.GetVersion(); v != "" {
		client.log = client.log.With(zap.String("v", v))
	}

	if opt.SessionStorage != nil {
		client.storage = &session.Loader{
			Storage: opt.SessionStorage,
		}
	}

	client.opts = mtproto.Options{
		PublicKeys:        opt.PublicKeys,
		Random:            opt.Random,
		Logger:            opt.Logger,
		AckBatchSize:      opt.AckBatchSize,
		AckInterval:       opt.AckInterval,
		RetryInterval:     opt.RetryInterval,
		MaxRetries:        opt.MaxRetries,
		CompressThreshold: opt.CompressThreshold,
		MessageID:         opt.MessageID,
		ExchangeTimeout:   opt.ExchangeTimeout,
		DialTimeout:       opt.DialTimeout,
		Clock:             opt.Clock,

		Types: getTypesMapping(),

		Tracer: client.tracer,
	}
	client.conn = client.createPrimaryConn(nil)

	return client
}

// init sets fields which needs explicit initialization, like maps or channels.
func (c *Client) init() {
	if c.domains == nil {
		c.domains = map[int]string{}
	}
	if c.cfg == nil {
		c.cfg = manager.NewAtomicConfig(tg.Config{})
	}
	c.ready = tdsync.NewResetReady()
	c.restart = make(chan struct{})
	c.migration = make(chan struct{}, 1)
	c.sessions = map[int]*pool.SyncSession{}
	c.subConns = map[int]CloseInvoker{}
	c.invoker = chainMiddlewares(InvokeFunc(c.invokeDirect), c.mw...)
	c.tg = tg.NewClient(c.invoker)
}
