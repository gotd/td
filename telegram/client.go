package telegram

import (
	"context"
	"errors"
	"io"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/clock"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
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
	Config() tg.Config
	Close() error
	Connect(ctx context.Context) error
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg *tg.Client

	connMux sync.Mutex
	connOpt mtproto.Options
	conn    clientConn

	trace tracer

	// Wrappers for external world, like current time, logs or PRNG.
	// Should be immutable.
	clock clock.Clock
	rand  io.Reader
	log   *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	appID   int    // immutable
	appHash string // immutable
	storage clientStorage

	updateHandler UpdateHandler // immutable
}

// NewClient creates new unstarted client.
func NewClient(appID int, appHash string, opt Options) *Client {
	// Set default values, if user does not set.
	opt.setDefaults()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		clock:         opt.Clock,
		rand:          opt.Random,
		log:           opt.Logger,
		ctx:           clientCtx,
		cancel:        clientCancel,
		appID:         appID,
		appHash:       appHash,
		updateHandler: opt.UpdateHandler,
	}

	if opt.SessionStorage != nil {
		client.storage = &session.Loader{
			Storage: opt.SessionStorage,
		}
	}

	client.connOpt = mtproto.Options{
		PublicKeys:    opt.PublicKeys,
		Transport:     opt.Transport,
		Network:       opt.Network,
		Random:        opt.Random,
		Logger:        opt.Logger,
		Handler:       client.handleMessage,
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
	client.conn = client.createConn(opt.Addr, connModeUpdates)

	// Initializing internal RPC caller.
	client.tg = tg.NewClient(client)

	return client
}

func (c *Client) restoreConnection(ctx context.Context) error {
	if c.storage == nil {
		return nil
	}
	data, err := c.storage.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		return nil
	}
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	// Re-initializing connection from persisted state.
	c.log.Info("Connection restored from state",
		zap.String("addr", data.Addr),
	)
	c.conn = c.createConn(data.Addr, connModeUpdates)

	return nil
}

// Connect initializes connection to Telegram server and starts internal
// read loop.
func (c *Client) Connect(ctx context.Context) error {
	if err := c.restoreConnection(ctx); err != nil {
		return xerrors.Errorf("restore: %w", err)
	}
	if err := c.conn.Connect(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleMessage(b *bin.Buffer) error {
	c.trace.OnMessage(b)
	return c.handleUpdates(b)
}
