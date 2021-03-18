package manager

import (
	"context"

	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

// SetupCallback is a optional setup connection callback.
type SetupCallback = func(ctx context.Context, invoker tg.Invoker) error

// ConnOptions is a Telegram client connection options.
type ConnOptions struct {
	DC      int
	Device  DeviceConfig
	Handler Handler
	Setup   SetupCallback
}

// SetDefaults sets default values.
func (c *ConnOptions) SetDefaults() {
	if c.DC == 0 {
		c.DC = 2
	}
	c.Device.SetDefaults()
	if c.Handler == nil {
		c.Handler = NoopHandler{}
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
	connOpts.SetDefaults()
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
	}

	conn.log = opts.Logger
	opts.Handler = conn
	opts.Logger = conn.log.Named("mtproto")
	conn.proto = mtproto.New(create, opts)

	return conn
}
