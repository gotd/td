package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
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
	Run(ctx context.Context) error
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}

// Client represents a MTProto client to Telegram.
type Client struct {
	// tg provides RPC calls via Client.
	tg      *tg.Client
	connOpt mtproto.Options
	conn    clientConn
	trace   tracer

	// Wrappers for external world, like logs or PRNG.
	// Should be immutable.
	rand io.Reader
	log  *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	appID   int    // immutable
	appHash string // immutable
	storage clientStorage
	ready   chan struct{}

	updateHandler UpdateHandler // immutable
}

func (c *Client) onMessage(b *bin.Buffer) error {
	return c.handleMessage(b)
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

	// Restoring persisted auth key.
	connOpt := c.connOpt
	var key crypto.AuthKeyWithID
	copy(key.AuthKey[:], data.AuthKey)
	copy(key.AuthKeyID[:], data.AuthKeyID)
	connOpt.Key = key
	connOpt.Salt = data.Salt
	c.connOpt = connOpt

	if connOpt.Key.AuthKey.ID() != connOpt.Key.AuthKeyID {
		return xerrors.New("corrupted key")
	}

	// Re-initializing connection from persisted state.
	c.log.Info("Connection restored from state",
		zap.String("addr", data.Addr),
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)
	c.conn = c.createConn(data.Addr, connModeUpdates)

	return nil
}

// Run starts client session and block until connection close.
// The f callback is called on successful session initialization and Run
// will return on f() result.
//
// Context of callback will be canceled if fatal error is detected.
func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) error {
	c.ready = make(chan struct{})
	if err := c.restoreConnection(ctx); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return c.conn.Run(gCtx)
	})
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return gCtx.Err()
		case <-c.ready:
			if err := f(gCtx); err != nil {
				return xerrors.Errorf("callback: %w", err)
			}
			// Should call cancel() to cancel gCtx.
			// This will terminate c.conn.Run().
			cancel()
			return nil
		}
	})
	if err := g.Wait(); !xerrors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (c *Client) handleMessage(b *bin.Buffer) error {
	c.trace.OnMessage(b)
	return c.handleUpdates(b)
}

func (c *Client) saveSession(addr string, cfg tg.Config, s mtproto.Session) error {
	if c.storage == nil {
		return nil
	}

	data, err := c.storage.Load(c.ctx)
	if errors.Is(err, session.ErrNotFound) {
		// Initializing new state.
		err = nil
		data = &session.Data{}
	}
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	// Updating previous data.
	data.Config = cfg
	data.AuthKey = s.Key.AuthKey[:]
	data.AuthKeyID = s.Key.AuthKeyID[:]
	data.DC = cfg.ThisDC
	data.Addr = addr
	data.Salt = s.Salt

	if err := c.storage.Save(c.ctx, data); err != nil {
		return xerrors.Errorf("save: %w", err)
	}

	c.log.Debug("Data saved",
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)
	return nil
}

func (c *Client) onSession(addr string, cfg tg.Config, s mtproto.Session) error {
	if err := c.saveSession(addr, cfg, s); err != nil {
		return xerrors.Errorf("save: %w", err)
	}
	close(c.ready)
	return nil
}

func (c *Client) createConn(addr string, mode connMode) clientConn {
	return newConn(
		c,
		addr,
		c.appID,
		mode,
		c.connOpt,
	)
}
