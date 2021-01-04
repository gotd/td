package telegram

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/clock"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

type protoConn interface {
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Connect(ctx context.Context) error
	Reconnect(ctx context.Context) error
	Session() mtproto.Session
	Close() error
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=connState
type connState byte

const (
	connCreated connState = iota
	connConnecting
	connConnected
	connInitializing
	connIdle
	connActive
	connReconnecting
	connClosing
	connClosed
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=connMode
type connMode byte

const (
	connModeUpdates connMode = iota
	connModeData
	connModeCDN
)

type conn struct {
	cfg     tg.Config
	appID   int
	appHash string
	mode    connMode
	state   connState
	proto   protoConn
	dc      int
	addr    string
	opt     mtproto.Options
	ongoing int
	clock   clock.Clock
	log     *zap.Logger
	latest  time.Time
	session *condOnce
	mux     sync.Mutex

	onMessage mtproto.Handler
	onSession onSessionHandler
}

func (c *conn) Config() tg.Config {
	if c == nil {
		return tg.Config{}
	}
	return c.cfg
}

func (c *conn) trackInvoke() (func(), error) {
	start := c.clock.Now()

	c.mux.Lock()
	defer c.mux.Unlock()

	if c.ongoing == 0 {
		if err := c.switchState(connActive); err != nil {
			return nil, err
		}
	}
	c.ongoing++
	c.latest = start

	return func() {
		c.mux.Lock()
		defer c.mux.Unlock()

		c.ongoing--
		if c.ongoing == 0 {
			_ = c.switchState(connIdle)
		}
		end := c.clock.Now()
		c.latest = end

		c.log.Debug("Invoke",
			zap.Duration("duration", end.Sub(start)),
			zap.Int("ongoing", c.ongoing),
		)
	}, nil
}

func (c *conn) switchState(next connState) error {
	if c == nil {
		return xerrors.New("nil conn")
	}
	if c.proto == nil {
		return xerrors.New("nil proto connection")
	}

	// TODO(ernado): implement FSM
	c.log.Info("State change",
		zap.Stringer("from", c.state),
		zap.Stringer("to", next),
	)

	c.state = next

	return nil
}

func (c *conn) connect(ctx context.Context) error {
	if err := c.switchState(connConnecting); err != nil {
		return xerrors.Errorf("state: %w", err)
	}
	if err := c.proto.Connect(ctx); err != nil {
		return xerrors.Errorf("connect: %w", err)
	}
	if err := c.switchState(connConnected); err != nil {
		return xerrors.Errorf("state: %w", err)
	}
	return nil
}

func (c *conn) Connect(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	if err := c.init(ctx); err != nil {
		return err
	}
	return nil
}

func (c *conn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// Tracking ongoing invokes.
	done, err := c.trackInvoke()
	if err != nil {
		return xerrors.Errorf("track: %w", err)
	}
	defer done()

	return c.proto.InvokeRaw(ctx, input, output)
}

func (c *conn) handleMessage(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}
	switch id {
	case mt.NewSessionCreatedTypeID:
		return c.handleSessionCreated(b)
	default:
		return c.onMessage(b)
	}
}

func (c *conn) init(ctx context.Context) error {
	if err := c.switchState(connInitializing); err != nil {
		return xerrors.Errorf("state: %w", err)
	}

	// TODO(ernado): Make versions configurable.
	const notAvailable = "n/a"

	q := proto.InitConnection{
		ID:             c.appID,
		SystemLangCode: "en",
		LangCode:       "en",
		SystemVersion:  notAvailable,
		DeviceModel:    notAvailable,
		AppVersion:     notAvailable,
		LangPack:       "",
		Query:          proto.GetConfig{},
	}
	var req bin.Object = proto.InvokeWithLayer{
		Layer: tg.Layer,
		Query: q,
	}
	if c.mode == connModeData || c.mode == connModeCDN {
		// This connection will not receive updates.
		req = proto.InvokeWithoutUpdates{
			Query: req,
		}
	}
	var response tg.Config

	if err := c.proto.InvokeRaw(ctx, req, &response); err != nil {
		return xerrors.Errorf("invoke: %w", err)
	}

	c.mux.Lock()
	// Now connection can be used for requests.
	c.latest = c.clock.Now()
	c.cfg = response
	onSessionErr := c.onSession(c.addr, c.cfg, c.proto.Session())
	c.mux.Unlock()

	if onSessionErr != nil {
		return xerrors.Errorf("onSession: %w", onSessionErr)
	}

	return nil
}

func (c *conn) Close() error {
	if err := c.switchState(connClosing); err != nil {
		return err
	}
	if err := c.proto.Close(); err != nil {
		return err
	}
	if err := c.switchState(connClosed); err != nil {
		return err
	}

	return nil
}

func (c *conn) handleSessionCreated(_ *bin.Buffer) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if err := c.switchState(connIdle); err != nil {
		return err
	}
	c.session.Done()

	// This should be persisted.
	_ = c.proto.Session()

	return nil
}

func (c *conn) reconnect(ctx context.Context) error {
	if err := c.proto.Reconnect(context.Background()); err != nil {
		_ = c.Close()
		return err
	}
	if err := c.init(ctx); err != nil {
		return err
	}
	return nil
}

func (c *conn) onReconnect() error {
	c.session.Reset()
	if err := c.switchState(connReconnecting); err != nil {
		return err
	}
	go func() {
		if err := c.reconnect(context.Background()); err != nil {
			c.log.Error("Failed to reconnect", zap.Error(err))
		}
	}()
	return nil
}

type onSessionHandler func(addr string, cfg tg.Config, session mtproto.Session) error

func newConn(
	addr string,
	appID int,
	appHash string,
	mode connMode,
	onSession onSessionHandler,
	opt mtproto.Options,
) *conn {
	c := &conn{
		appID:     appID,
		appHash:   appHash,
		mode:      mode,
		dc:        0,
		addr:      addr,
		opt:       opt,
		clock:     opt.Clock,
		log:       opt.Logger.Named("conn"),
		onMessage: opt.Handler,
		session:   createCondOnce(),
		onSession: onSession,
	}
	c.opt.Handler = c.handleMessage
	c.opt.OnReconnect = c.onReconnect
	c.proto = mtproto.NewConn(c.addr, c.opt)
	return c
}
