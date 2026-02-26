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
	onDead func(error)

	// Wrappers for external world, like logs or PRNG.
	// Should be immutable.
	clock clock.Clock // immutable
	log   *zap.Logger // immutable

	// Handler passed by client.
	handler Handler // immutable

	// State fields.
	cfg tg.Config
	// pending buffers OnSession events until initConnection config is available.
	pending []mtproto.Session
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

	// Quote (PFS): "Once auth.bindTempAuthKey has been executed successfully,
	// the client can continue generating API calls as usual."
	// Link: https://core.telegram.org/api/pfs
	//
	// In PFS mode bind is performed before initConnection, so OnSession can happen
	// before config is ready. We must not block read-loop handler here.
	c.mux.Lock()
	c.pending = append(c.pending, session)
	c.mux.Unlock()

	if !c.configReady() {
		return nil
	}
	return c.flushPendingSession()
}

func (c *Conn) configReady() bool {
	select {
	case <-c.gotConfig.Ready():
		return true
	default:
		return false
	}
}

func (c *Conn) flushPendingSession() error {
	c.mux.Lock()
	pending := append([]mtproto.Session(nil), c.pending...)
	cfg := c.cfg
	c.pending = c.pending[:0]
	c.mux.Unlock()
	if len(pending) == 0 {
		return nil
	}
	for _, s := range pending {
		// Preserve event ordering in case multiple session events arrive before
		// config is ready.
		if err := c.handler.OnSession(cfg, s); err != nil {
			return err
		}
	}
	return nil
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
				c.onDead(err)
			}
		}
	}()
	return c.proto.Run(ctx, func(ctx context.Context) error {
		// Signal death on init error to unblock waiters in waitSession/OnSession.
		err := c.init(ctx)
		if err != nil {
			c.dead.Signal()
		}
		return err
	})
}

func (c *Conn) waitSession(ctx context.Context) error {
	select {
	// Connection is considered ready only after initConnection succeeded.
	case <-c.gotConfig.Ready():
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
	// Pool should expose readiness only when Invoke can send API calls.
	return c.gotConfig.Ready()
}

// Invoke implements Invoker.
func (c *Conn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// Tracking ongoing invokes.
	defer c.trackInvoke()()
	if err := c.waitSession(ctx); err != nil {
		return errors.Wrap(err, "waitSession")
	}

	q := c.wrapRequest(noopDecoder{input})
	req := c.wrapRequest(&tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	})

	return c.proto.Invoke(ctx, req, output)
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
	return c.flushPendingSession()
}

// Ping calls ping for underlying protocol connection.
func (c *Conn) Ping(ctx context.Context) error {
	return c.proto.Ping(ctx)
}
