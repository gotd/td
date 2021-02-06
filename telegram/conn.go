package telegram

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/internal/tdsync"
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
	// Connection parameters.
	addr string   // immutable
	mode connMode // immutable
	// MTProto connection.
	proto protoConn // immutable

	// InitConnection parameters.
	appID  int          // immutable
	device DeviceConfig // immutable

	// Wrappers for external world, like logs or PRNG.
	// Should be immutable.
	clock clock.Clock // immutable
	log   *zap.Logger // immutable

	// Handler passed by client.
	handler connHandler // immutable

	// State fields.
	cfg     tg.Config
	ongoing int
	latest  time.Time
	mux     sync.Mutex

	sessionInit *tdsync.Ready // immutable
	gotConfig   *tdsync.Ready // immutable
	dead        *tdsync.Ready // immutable
}

func (c *conn) OnSession(session mtproto.Session) error {
	c.log.Info("SessionInit")
	c.sessionInit.Signal()

	// Waiting for config, because OnSession can occur before we set config.
	select {
	case <-c.gotConfig.Ready():
	case <-c.dead.Ready():
	}

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

func (c *conn) Run(ctx context.Context) (err error) {
	defer c.dead.Signal()
	defer func() {
		c.log.Info("Connection dead", zap.Error(err))
	}()
	return c.proto.Run(ctx, c.init)
}

func (c *conn) waitSession(ctx context.Context) error {
	select {
	case <-c.sessionInit.Ready():
		return nil
	case <-c.dead.Ready():
		return pool.ErrConnDead
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *conn) Ready() <-chan struct{} {
	return c.sessionInit.Ready()
}

func (c *conn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// Tracking ongoing invokes.
	defer c.trackInvoke()()
	if err := c.waitSession(ctx); err != nil {
		return xerrors.Errorf("waitSession: %w", err)
	}

	return c.proto.InvokeRaw(ctx, c.wrapRequest(noopDecoder{input}), output)
}

func (c *conn) OnMessage(b *bin.Buffer) error {
	return c.handler.onMessage(b)
}

type noopDecoder struct {
	bin.Encoder
}

func (n noopDecoder) Decode(b *bin.Buffer) error {
	panic("implement me")
}

func (c *conn) wrapRequest(req bin.Object) bin.Object {
	if c.mode != connModeUpdates {
		return &tg.InvokeWithoutUpdatesRequest{
			Query: req,
		}
	}

	return req
}

func (c *conn) init(ctx context.Context) error {
	defer c.gotConfig.Signal()
	c.log.Debug("Initializing")

	q := c.wrapRequest(&tg.InitConnectionRequest{
		APIID:          c.appID,
		DeviceModel:    c.device.DeviceModel,
		SystemVersion:  c.device.SystemVersion,
		AppVersion:     c.device.AppVersion,
		SystemLangCode: c.device.SystemLangCode,
		LangPack:       c.device.LangPack,
		LangCode:       c.device.LangCode,
		Query:          c.wrapRequest(&tg.HelpGetConfigRequest{}),
	})
	req := c.wrapRequest(&tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	})

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
