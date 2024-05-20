package manager

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
)

// SetupCallback is an optional setup connection callback.
type SetupCallback = func(ctx context.Context, invoker tg.Invoker) error

// ConnOptions is a Telegram client connection options.
type ConnOptions struct {
	DC      int
	Test    bool
	Device  DeviceConfig
	Handler Handler
	Setup   SetupCallback
	OnDead  func()
	Backoff func(ctx context.Context) backoff.BackOff
}

func defaultBackoff(c clock.Clock) func(ctx context.Context) backoff.BackOff {
	return func(ctx context.Context) backoff.BackOff {
		b := backoff.NewExponentialBackOff()
		b.Clock = c
		b.MaxElapsedTime = time.Second * 30
		b.MaxInterval = time.Second * 5
		return backoff.WithContext(b, ctx)
	}
}

// setDefaults sets default values.
func (c *ConnOptions) setDefaults(connClock clock.Clock) {
	if c.DC == 0 {
		c.DC = 2
	}
	// It's okay to use zero value Test.
	c.Device.SetDefaults()
	if c.Handler == nil {
		c.Handler = NoopHandler{}
	}
	if c.Backoff == nil {
		c.Backoff = defaultBackoff(connClock)
	}
}

// CreateConn creates new connection.
func CreateConn(
	create mtproto.Dialer,
	mode ConnMode,
	appID int,
	opts mtproto.Options,
	connOpts ConnOptions,
) *Conn {
	connOpts.setDefaults(opts.Clock)
	conn := &Conn{
		mode:        mode,
		appID:       appID,
		device:      connOpts.Device,
		clock:       opts.Clock,
		handler:     connOpts.Handler,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
		setup:       connOpts.Setup,
		onDead:      connOpts.OnDead,
		connBackoff: connOpts.Backoff,
	}

	conn.log = opts.Logger
	opts.DC = connOpts.DC
	if connOpts.Test {
		// New key exchange algorithm requires DC ID and uses mapping like MTProxy.
		// +10000 for test DC, *-1 for media-only.
		opts.DC += 10000
	}
	opts.Handler = conn
	opts.Logger = conn.log.Named("mtproto")
	conn.proto = mtproto.New(create, opts)

	return conn
}
