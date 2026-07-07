package manager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/gotd/log"

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
	dc   int      // immutable
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
	log   log.Helper  // immutable

	// Handler passed by client.
	handler Handler // immutable

	// State fields.
	cfg tg.Config
	// cdnNeedsInit mirrors TDesktop connectionInited state for CDN transport.
	// true means requests must go via invokeWithLayer(initConnection).
	cdnNeedsInit atomic.Bool
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
	c.log.Info(context.Background(), "SessionInit")
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

		c.log.Debug(context.Background(), "Invoke",
			log.Duration("duration", end.Sub(start)),
			log.Int("ongoing", c.ongoing),
		)
	}
}

// Run initialize connection.
func (c *Conn) Run(ctx context.Context) (err error) {
	defer c.dead.Signal()
	defer func() {
		if err != nil && ctx.Err() == nil {
			c.log.Debug(ctx, "Connection dead", log.Error(err))
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
	// Fast path: config already available, do not log.
	select {
	case <-c.gotConfig.Ready():
		return nil
	default:
	}

	// Connection not ready yet. A request blocked here while updates keep
	// flowing means the connection's read loop is alive but init
	// (initConnection/help.getConfig) has not completed — the exact shape of
	// "can receive updates but cannot issue requests".
	start := c.clock.Now()
	c.log.Debug(ctx, "Invoke waiting for connection to become ready")
	select {
	// Connection is considered ready only after mode-specific init succeeded.
	case <-c.gotConfig.Ready():
		c.log.Debug(ctx, "Connection became ready", log.Duration("waited", c.clock.Now().Sub(start)))
		return nil
	case <-c.dead.Ready():
		c.log.Debug(ctx, "Connection died while waiting for readiness", log.Duration("waited", c.clock.Now().Sub(start)))
		return pool.ErrConnDead
	case <-ctx.Done():
		c.log.Debug(ctx, "Context done while waiting for connection readiness",
			log.Duration("waited", c.clock.Now().Sub(start)),
			log.Error(ctx.Err()),
		)
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

	if c.mode == ConnModeCDN {
		// CDN mode has dedicated request wrapping rules (see invokeCDN).
		err := c.invokeCDN(ctx, input, output)
		return err
	}
	q := c.wrapRequest(noopDecoder{input})
	req := c.wrapRequest(&tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	})
	err := c.proto.Invoke(ctx, req, output)
	return err
}
func (c *Conn) invokeCDN(
	ctx context.Context,
	input bin.Encoder,
	output bin.Decoder,
) error {
	// TDesktop model:
	// - while connection is "not inited": wrap every query in invokeWithLayer(initConnection);
	// - after first successful reply: use raw CDN methods;
	// - if server returns CONNECTION_NOT_INITED/LAYER_INVALID on raw call:
	//   mark "not inited" and retry wrapped once.
	if c.cdnNeedsInit.Load() {
		err := c.invokeCDNWrapped(ctx, input, output)
		if err == nil {
			c.cdnNeedsInit.Store(false)
			return nil
		}
		return err
	}

	err := c.invokeCDNRaw(ctx, input, output)
	if err == nil {
		return nil
	}
	if c.shouldCDNRetryWrapped(err) {
		c.cdnNeedsInit.Store(true)
		retryErr := c.invokeCDNWrapped(ctx, input, output)
		if retryErr == nil {
			c.cdnNeedsInit.Store(false)
			return nil
		}
		return retryErr
	}
	return err
}
func (c *Conn) invokeCDNWrapped(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	req := &tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: c.cdnInitRequest(noopDecoder{input}),
	}
	return c.proto.Invoke(ctx, req, output)
}
func (c *Conn) invokeCDNRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return c.proto.Invoke(ctx, input, output)
}
func (c *Conn) shouldCDNRetryWrapped(err error) bool {
	if err == nil {
		return false
	}
	if rpcErr, ok := tgerr.As(err); ok {
		// Retry wrapped only for not-inited/layer-invalid transport state.
		v := rpcErr.IsOneOf(
			"CONNECTION_NOT_INITED",
			"CONNECTION_LAYER_INVALID",
		)
		return v
	}
	return false
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
	if c.mode == ConnModeData {
		return &tg.InvokeWithoutUpdatesRequest{
			Query: req,
		}
	}

	return req
}

func (c *Conn) cdnInitRequest(query bin.Object) bin.Object {
	// Match TDesktop CDN init wrapper:
	// only device/system are anonymized, the rest of initConnection
	// parameters stay aligned with regular connection settings.
	return &tg.InitConnectionRequest{
		APIID:          c.appID,
		DeviceModel:    "n/a",
		SystemVersion:  "n/a",
		AppVersion:     c.device.AppVersion,
		SystemLangCode: c.device.SystemLangCode,
		LangPack:       c.device.LangPack,
		LangCode:       c.device.LangCode,
		Proxy:          c.device.Proxy,
		Params:         c.device.Params,
		Query:          query,
	}
}
func (c *Conn) init(ctx context.Context) error {
	c.log.Debug(ctx, "Initializing")

	if c.mode == ConnModeCDN {
		// CDN connections skip help.getConfig init flow and become ready
		// immediately after MTProto auth-key exchange.
		c.cdnNeedsInit.Store(true)
		c.mux.Lock()
		c.latest = c.clock.Now()
		c.cfg = tg.Config{ThisDC: c.dc}
		c.mux.Unlock()
		c.gotConfig.Signal()
		err := c.flushPendingSession()
		return err
	}
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
		c.log.Debug(ctx, "Retrying connection initialization",
			log.Error(err), log.Duration("duration", duration),
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

	c.log.Debug(ctx, "Connection initialized, ready to invoke", log.Int("this_dc", cfg.ThisDC))
	c.gotConfig.Signal()
	err := c.flushPendingSession()
	return err
}

// Ping calls ping for underlying protocol connection.
func (c *Conn) Ping(ctx context.Context) error {
	return c.proto.Ping(ctx)
}
