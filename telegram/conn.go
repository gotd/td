package telegram

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

type protoConn interface {
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Run(ctx context.Context, f func(ctx context.Context) error) error
}

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
	mode    connMode
	proto   protoConn
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

func (c *conn) trackInvoke() func() {
	start := c.clock.Now()

	c.mux.Lock()
	defer c.mux.Unlock()

	c.ongoing++
	c.latest = start

	return func() {
		c.mux.Lock()
		defer c.mux.Unlock()

		c.ongoing--
		end := c.clock.Now()
		c.latest = end

		c.log.Debug("Invoke",
			zap.Duration("duration", end.Sub(start)),
			zap.Int("ongoing", c.ongoing),
		)
	}
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
	defer c.trackInvoke()()
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
	// TODO(ernado): Make versions configurable.
	const notAvailable = "n/a"

	q := &tg.InitConnectionRequest{
		APIID:          c.appID,
		SystemLangCode: "en",
		LangCode:       "en",
		SystemVersion:  notAvailable,
		DeviceModel:    notAvailable,
		AppVersion:     notAvailable,
		LangPack:       "",
		Query:          &tg.HelpGetConfigRequest{},
	}
	var req bin.Object = &tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	}
	if c.mode == connModeData || c.mode == connModeCDN {
		// This connection will not receive updates.
		req = &tg.InvokeWithoutUpdatesRequest{
			Query: req,
		}
	}

	var cfg tg.Config
	if err := c.proto.InvokeRaw(ctx, req, &cfg); err != nil {
		return xerrors.Errorf("invoke: %w", err)
	}

	c.mux.Lock()
	c.latest = c.clock.Now()
	c.cfg = cfg
	c.mux.Unlock()

	return nil
}

func newConn(
	handler connHandler,
	addr string,
	appID int,
	mode connMode,
	opt mtproto.Options,
) *conn {
	c := &conn{
		appID:       appID,
		mode:        mode,
		addr:        addr,
		clock:       opt.Clock,
		log:         opt.Logger.Named("conn"),
		handler:     handler,
		sessionInit: make(chan struct{}),
	}
	opt.Handler = c
	c.proto = mtproto.New(addr, opt)
	return c
}
