package telegram

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

type protoConn interface {
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Run(ctx context.Context, f func(ctx context.Context) error) error
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

type connHandler interface {
	onSession(addr string, cfg tg.Config, s mtproto.Session) error
	onMessage(b *bin.Buffer) error
}

type conn struct {
	addr    string
	cfg     tg.Config
	appID   int
	appHash string
	mode    connMode
	state   connState
	proto   protoConn
	opt     mtproto.Options
	ongoing int
	clock   clock.Clock
	log     *zap.Logger
	latest  time.Time
	mux     sync.Mutex

	handler connHandler

	sessionInitOnce sync.Once
	sessionInit     chan struct{}
}

func (c *conn) OnSession(session mtproto.Session) error {
	c.sessionInitOnce.Do(func() {
		close(c.sessionInit)
	})

	c.mux.Lock()
	cfg := c.cfg
	c.mux.Unlock()

	return c.handler.onSession(c.addr, cfg, session)
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
	c.log.Debug("State change",
		zap.Stringer("from", c.state),
		zap.Stringer("to", next),
	)

	c.state = next

	return nil
}

func (c *conn) Run(ctx context.Context) error {
	return c.proto.Run(ctx, c.init)
}

func (c *conn) waitSession(ctx context.Context) error {
	select {
	case <-c.sessionInit:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *conn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// Tracking ongoing invokes.
	done, err := c.trackInvoke()
	if err != nil {
		return xerrors.Errorf("track: %w", err)
	}
	defer done()
	if err := c.waitSession(ctx); err != nil {
		return xerrors.Errorf("waitSession: %w", err)
	}

	return c.proto.InvokeRaw(ctx, input, output)
}

func (c *conn) OnMessage(b *bin.Buffer) error {
	return c.handler.onMessage(b)
}

func (c *conn) init(ctx context.Context) error {
	c.log.Debug("Initializing")
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
	c.mux.Unlock()

	return nil
}

func newConn(
	handler connHandler,
	addr string,
	appID int,
	appHash string,
	mode connMode,
	opt mtproto.Options,
) *conn {
	c := &conn{
		appID:       appID,
		appHash:     appHash,
		mode:        mode,
		addr:        addr,
		opt:         opt,
		clock:       opt.Clock,
		log:         opt.Logger.Named("conn"),
		handler:     handler,
		sessionInit: make(chan struct{}),
	}
	c.opt.Handler = c
	c.proto = mtproto.New(c.addr, c.opt)
	return c
}
