package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/dcmanager"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/lifetime"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/proto"
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

type invoker interface {
	Run(ctx context.Context, f func(context.Context) error) error
	InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg *tg.Client // immutable

	dcm invoker

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
		rand:          opt.Random,
		log:           opt.Logger,
		ctx:           clientCtx,
		cancel:        clientCancel,
		appID:         appID,
		appHash:       appHash,
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

	client.dcm = dcmanager.New(appID, dcmanager.Options{
		Addr:          opt.Addr,
		UpdateHandler: client.handleUpdates,
		ConfigSaver:   client.saveConfig,
		MTOptions: mtproto.Options{
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
		Logger: opt.Logger.Named("dc_manager"),
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
	if dcm, ok := c.dcm.(interface {
		UpdateConfig(cfg dcmanager.Config)
	}); ok && c.storage != nil {
		sess, err := c.storage.Load(ctx)
		if err != nil {
			if !errors.Is(err, session.ErrNotFound) {
				return err
			}
		} else {
			dcm.UpdateConfig(dcmanager.Config{
				TGConfig:  sess.Config,
				PrimaryDC: sess.DC,
				AuthKey: crypto.AuthKey{
					Value: sess.AuthKey,
					ID:    sess.AuthKeyID,
				},
				Salt: sess.Salt,
			})
		}
	}

	select {
	case <-c.ctx.Done():
		return xerrors.Errorf("client already closed: %w", c.ctx.Err())
	default:
	}

	c.log.Info("Starting")
	defer c.log.Info("Closed")

	life, err := lifetime.Start(c.dcm)
	if err != nil {
		return err
	}
	defer life.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	echan := make(chan error)
	go func() { echan <- f(ctx) }()
	go func() { echan <- life.Wait() }()
	return <-echan
}

// InvokeRaw sens input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Client) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return c.dcm.InvokeRaw(ctx, input, output)
}

func (c *Client) saveConfig(cfg dcmanager.Config) error {
	if c.storage == nil {
		return nil
	}

	if err := c.storage.Save(c.ctx, &session.Data{
		Config:    cfg.TGConfig,
		DC:        cfg.PrimaryDC,
		Addr:      "",
		AuthKey:   cfg.AuthKey.Value,
		AuthKeyID: cfg.AuthKey.ID,
		Salt:      cfg.Salt,
	}); err != nil {
		return xerrors.Errorf("save: %w", err)
	}

	c.log.Debug("Data saved",
		zap.String("key_id", fmt.Sprintf("%x", cfg.AuthKey.ID)),
	)
	return nil
}
