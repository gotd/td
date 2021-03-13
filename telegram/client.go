package telegram

import (
	"context"
	"errors"
	"io"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/lifetime"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
)

// UpdateHandler will be called on received updates from Telegram.
type UpdateHandler interface {
	Handle(ctx context.Context, u *tg.Updates) error
	HandleShort(ctx context.Context, u *tg.UpdateShort) error
}

// Available MTProto default server addresses.
//
// See https://my.telegram.org/apps.
const (
	AddrProduction = "149.154.167.50:443"
	AddrTest       = "149.154.167.40:443"
)

// Port is default port used by telegram.
const Port = 443

// Test-only credentials. Can be used with AddrTest and TestAuth to
// test authentication.
//
// Reference:
//	* https://github.com/telegramdesktop/tdesktop/blob/5f665b8ecb48802cd13cfb48ec834b946459274a/docs/api_credentials.md
const (
	TestAppID   = 17349
	TestAppHash = "344583e45741c457fe1862106095a5eb"
)

type clientStorage interface {
	Load(ctx context.Context) (*session.Data, error)
	Save(ctx context.Context, data *session.Data) error
}

type conn interface {
	Run(ctx context.Context, f func(context.Context) error) error
	InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	tg        *tg.Client
	primary   conn
	primaryDC int
	pmux      sync.RWMutex
	migrateOp *tdsync.SinglePerformer

	sess    mtproto.Session
	cfg     tg.Config
	dataMux sync.RWMutex

	others map[int]conn
	omux   sync.RWMutex
	lf     *lifetime.Manager

	// Wrappers for external world, like logs or PRNG.
	rand  io.Reader   // immutable
	log   *zap.Logger // immutable
	clock clock.Clock // immutable

	// Client context. Will be canceled by Run on exit.
	ctx    context.Context    // immutable
	cancel context.CancelFunc // immutable

	// Client config.
	appID   int             // immutable
	appHash string          // immutable
	addr    string          // immutable
	device  DeviceConfig    // immutable
	opts    mtproto.Options // immutable

	// Session storage.
	storage clientStorage // immutable, nillable

	// Telegram updates handler.
	updateHandler UpdateHandler // immutable
}

// NewClient creates new unstarted client.
func NewClient(appID int, appHash string, opt Options) *Client {
	// Set default values, if user does not set.
	opt.setDefaults()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		primaryDC: opt.DC,
		migrateOp: &tdsync.SinglePerformer{},
		others:    map[int]conn{},
		lf:        lifetime.NewManager(),

		rand:  opt.Random,
		log:   opt.Logger,
		clock: opt.Clock,

		ctx:    clientCtx,
		cancel: clientCancel,

		appID:   appID,
		appHash: appHash,
		addr:    opt.Addr,
		device:  opt.Device,

		opts: mtproto.Options{
			PublicKeys:    opt.PublicKeys,
			Transport:     opt.Transport,
			Network:       opt.Network,
			Random:        opt.Random,
			Logger:        opt.Logger,
			AckBatchSize:  opt.AckBatchSize,
			AckInterval:   opt.AckInterval,
			RetryInterval: opt.RetryInterval,
			MaxRetries:    opt.MaxRetries,
			MessageID:     opt.MessageID,
			Clock:         opt.Clock,

			Types: tmap.New(
				tg.TypesMap(),
				mt.TypesMap(),
				proto.TypesMap(),
			),
		},
		updateHandler: opt.UpdateHandler,
	}

	// Including version into client logger to help with debugging.
	if v := getVersion(); v != "" {
		client.log = client.log.With(zap.String("v", v))
	}

	if opt.SessionStorage != nil {
		client.storage = &session.Loader{
			Storage: opt.SessionStorage,
		}
	}

	client.tg = tg.NewClient(client)

	return client
}

// Run starts client session and block until connection close.
// The f callback is called on successful session initialization and Run
// will return on f() result.
//
// Context of callback will be canceled if fatal error is detected.
func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) error {
	select {
	case <-c.ctx.Done():
		return xerrors.Errorf("client already closed: %w", c.ctx.Err())
	default:
	}

	var (
		dcInfo     tg.DCOption
		reuseCreds bool
	)

	// Try to load previous session.
	if err := c.storageLoad(ctx); err != nil {
		// Something bad happened with storage.
		if !errors.Is(err, session.ErrNotFound) {
			c.log.Error("Storage failure", zap.Error(err))
			return err
		}

		c.log.Info("Session not found. Using server provided in opts as primary DC.",
			zap.Int("dc_id", c.primaryDC),
			zap.String("dc_addr", c.addr),
		)

		dcInfo, err = c.primaryDCOption()
		if err != nil {
			return err
		}
	} else {
		c.log.Info("Previous session restored from storage.",
			zap.Int("dc_id", c.primaryDC),
			zap.String("dc_addr", c.addr),
		)

		dcInfo, err = c.lookupDC(c.primaryDC)
		if err != nil {
			return err
		}

		reuseCreds = true
	}

	if err := c.connectPrimary(ctx, dcInfo, reuseCreds); err != nil {
		return err
	}

	c.log.Info("Started")
	defer c.log.Info("Closed")

	// Close client context on exit.
	defer c.cancel()

	// Close connections on exit.
	defer c.lf.Close()

	// Close 'f' ctx on exit.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e := make(chan error)
	go func() { e <- f(ctx) }()
	go func() { e <- c.lf.Wait() }()
	return <-e
}
