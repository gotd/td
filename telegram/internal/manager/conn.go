package manager

import (
	"context"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type protoConn interface {
	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Run(ctx context.Context, f func(ctx context.Context) error) error
	Ping(ctx context.Context) error
}

//go:generate go run -modfile=../../../_tools/go.mod golang.org/x/tools/cmd/stringer -type=ConnMode

// ConnMode represents connection mode.
type ConnMode byte

const (
	// ConnModeUpdates is update connection mode.
	ConnModeUpdates ConnMode = iota
	// ConnModeData is data connection mode.
	ConnModeData
	// ConnModeCDN is CDN connection mode.
	ConnModeCDN
)

// Conn is a Telegram client connection.
type Conn struct {
	// Connection parameters.
	mode ConnMode // immutable
	// MTProto connection.
	proto protoConn // immutable

	// InitConnection parameters.
	appID  int          // immutable
	device DeviceConfig // immutable

	// setup is callback which called after initConnection, but before ready signaling.
	// This is necessary to transfer auth from previous connection to another DC.
	setup SetupCallback // nilable

	// onDead is called on connection death.
	onDead func()

	// Wrappers for external world, like logs or PRNG.
	// Should be immutable.
	clock clock.Clock // immutable
	log   *zap.Logger // immutable

	// Handler passed by client.
	handler Handler // immutable

	// State fields.
	cfg     tg.Config
	ongoing int
	latest  time.Time
	mux     sync.Mutex

	sessionInit *tdsync.Ready // immutable
	gotConfig   *tdsync.Ready // immutable
	dead        *tdsync.Ready // immutable

	connBackoff func(ctx context.Context) backoff.BackOff // immutable
}

// OnSession implements mtproto.Handler.
func (c *Conn) OnSession(session mtproto.Session) error {
	c.log.Info("SessionInit")
	c.sessionInit.Signal()

	// Waiting for config, because OnSession can occur before we set config.
	select {
	case <-c.gotConfig.Ready():
	case <-c.dead.Ready():
		return nil
	}

	c.mux.Lock()
	cfg := c.cfg
	c.mux.Unlock()

	return c.handler.OnSession(cfg, session)
}

func (c *Conn) trackInvoke() func() {
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

// Run initialize connection.
func (c *Conn) Run(ctx context.Context) (err error) {
	defer c.dead.Signal()
	defer func() {
		if err != nil && ctx.Err() == nil {
			c.log.Debug("Connection dead", zap.Error(err))
			if c.onDead != nil {
				c.onDead()
			}
		}
	}()
	return c.proto.Run(ctx, func(ctx context.Context) error {
		// Signal death on init error. Otherwise connection shutdown
		// deadlocks in OnSession that occurs before init fails.
		err := c.init(ctx)
		if err != nil {
			c.dead.Signal()
		}
		return err
	})
}

func (c *Conn) waitSession(ctx context.Context) error {
	select {
	case <-c.sessionInit.Ready():
		return nil
	case <-c.dead.Ready():
		return pool.ErrConnDead
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Ready returns channel to determine connection readiness.
// Useful for pooling.
func (c *Conn) Ready() <-chan struct{} {
	return c.sessionInit.Ready()
}

// Invoke implements Invoker.
func (c *Conn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// Tracking ongoing invokes.
	defer c.trackInvoke()()
	if err := c.waitSession(ctx); err != nil {
		return errors.Wrap(err, "waitSession")
	}

	return c.proto.Invoke(ctx, c.wrapRequest(noopDecoder{input}), output)
}

// OnMessage implements mtproto.Handler.
func (c *Conn) OnMessage(b *bin.Buffer) error {
	return c.handler.OnMessage(b)
}

type noopDecoder struct {
	bin.Encoder
}

func (n noopDecoder) Decode(b *bin.Buffer) error {
	return errors.New("not implemented")
}

func (c *Conn) wrapRequest(req bin.Object) bin.Object {
	if c.mode != ConnModeUpdates {
		return &tg.InvokeWithoutUpdatesRequest{
			Query: req,
		}
	}

	return req
}

func (c *Conn) init(ctx context.Context) error {
	c.log.Debug("Initializing")

	q := c.wrapRequest(&tg.InitConnectionRequest{
		APIID:          c.appID,
		DeviceModel:    c.device.DeviceModel,
		SystemVersion:  c.device.SystemVersion,
		AppVersion:     c.device.AppVersion,
		SystemLangCode: c.device.SystemLangCode,
		LangPack:       c.device.LangPack,
		LangCode:       c.device.LangCode,
		Proxy:          c.device.Proxy,
		Params:         c.device.Params,
		Query:          c.wrapRequest(&tg.HelpGetConfigRequest{}),
	})
	req := c.wrapRequest(&tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	})

	var cfg tg.Config
	if err := backoff.RetryNotify(func() error {
		if err := c.proto.Invoke(ctx, req, &cfg); err != nil {
			if tgerr.Is(err, tgerr.ErrFloodWait) {
				// Server sometimes returns FLOOD_WAIT(0) if you create
				// multiple connections in short period of time.
				//
				// See https://github.com/gotd/td/issues/388.
				return errors.Wrap(err, "flood wait")
			}
			// Not retrying other errors.
			return backoff.Permanent(errors.Wrap(err, "invoke"))
		}

		return nil
	}, c.connBackoff(ctx), func(err error, duration time.Duration) {
		c.log.Debug("Retrying connection initialization",
			zap.Error(err), zap.Duration("duration", duration),
		)
	}); err != nil {
		return errors.Wrap(err, "initConnection")
	}

	if c.setup != nil {
		if err := c.setup(ctx, c); err != nil {
			return errors.Wrap(err, "setup")
		}
	}

	c.mux.Lock()
	c.latest = c.clock.Now()
	c.cfg = cfg
	c.mux.Unlock()

	c.gotConfig.Signal()
	return nil
}

// Ping calls ping for underlying protocol connection.
func (c *Conn) Ping(ctx context.Context) error {
	return c.proto.Ping(ctx)
}
