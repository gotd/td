package manager

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/mtproto"
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
	id int64,
	mode ConnMode,
	appID int,
	addr string,
	opts mtproto.Options,
	connOpts ConnOptions,
) *Conn {
	connOpts.SetDefaults()
	conn := &Conn{
		id:          id,
		addr:        addr,
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

	conn.log = opts.Logger.Named("conn").With(
		zap.Int64("conn_id", conn.id),
		zap.Int("dc_id", connOpts.DC),
	)
	opts.Handler = conn
	opts.Logger = conn.log.Named("mtproto").With(zap.String("addr", conn.addr))
	conn.proto = mtproto.New(strconv.Itoa(connOpts.DC)+"|"+conn.addr, opts)

	return conn
}
