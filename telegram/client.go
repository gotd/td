package telegram

import (
	"context"
	"io"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/mtproto"
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

type clientConn interface {
	Run(ctx context.Context) error
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg *tg.Client // immutable

	// Telegram device information.
	device DeviceConfig // immutable

	// Connection pool.
	conn clientConn

	// Wrappers for external world, like logs or PRNG.
	rand  io.Reader   // immutable
	log   *zap.Logger // immutable
	clock clock.Clock // immutable

	// Client context. Will be canceled by Run on exit.
	ctx    context.Context    // immutable
	cancel context.CancelFunc // immutable

	// Client config.
	appID   int    // immutable
	appHash string // immutable
	// Session storage.
	storage clientStorage // immutable,nilable

	// Ready signal channel, sends signal when client connection is ready.
	// Resets on reconnect.
	ready *tdsync.ResetReady // immutable

	// Telegram updates handler.
	updateHandler UpdateHandler // immutable
}

// OnSession implements ConnHandler.
func (c *Client) OnSession(addr string, cfg tg.Config, s mtproto.Session) error {
	return nil
}

// OnMessage implements ConnHandler.
func (c *Client) OnMessage(b *bin.Buffer) error {
	return c.handleUpdates(b)
}

// getVersion optimistically gets current client version.
//
// Does not handle replace directives.
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// Hard-coded package name. Probably we can generate this via parsing
	// the go.mod file.
	const pkg = "github.com/gotd/td"
	for _, d := range info.Deps {
		if strings.HasPrefix(d.Path, pkg) {
			return d.Version
		}
	}
	return ""
}

// Port is default port used by telegram.
const Port = 443

// NewClient creates new unstarted client.
func NewClient(appID int, appHash string, opt Options) *Client {
	// Set default values, if user does not set.
	opt.setDefaults()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		rand:          opt.Random,
		log:           opt.Logger,
		ctx:           clientCtx,
		cancel:        clientCancel,
		appID:         appID,
		appHash:       appHash,
		updateHandler: opt.UpdateHandler,
		clock:         opt.Clock,
		device:        opt.Device,
		ready:         tdsync.NewResetReady(),
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

	options := mtproto.Options{
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
	}
	client.conn = pool.NewPool(appID, client, pool.Options{
		Addr:           opt.Addr,
		Device:         pool.DeviceConfig(opt.Device),
		SessionStorage: client.storage,
		Logger:         opt.Logger.Named("pool"),
		MTProto:        options,
	})

	// Initializing internal RPC caller.
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
		return xerrors.Errorf("client already closed: %w", ctx.Err())
	default:
	}

	c.log.Info("Starting")
	defer c.log.Info("Closed")
	// Cancel client on exit.
	defer c.cancel()

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.conn.Run)
	g.Go(func(gCtx context.Context) error {
		if err := f(gCtx); err != nil {
			return xerrors.Errorf("callback: %w", err)
		}
		// Should call cancel() to cancel gCtx.
		// This will terminate c.conn.Run().
		c.log.Debug("Callback returned, stopping")
		g.Cancel()
		return nil
	})
	if err := g.Wait(); !xerrors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
