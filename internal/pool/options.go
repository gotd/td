package pool

import (
	"go.uber.org/zap"

	"github.com/gotd/td/mtproto"
)

// Options is a Pool type options.
type Options struct {
	// Addr to connect.
	//
	// If not provided, AddrProduction will be used by default.
	Addr string
	// DC ID of given Addr.
	//
	//	If not provided, DC 2 will be used by default.
	ID int
	// Telegram device information.
	Device DeviceConfig
	// SessionStorage will be used to load and save session data.
	// NB: Very sensitive data, save with care.
	SessionStorage SessionStorage
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// MTProto options for connections.
	MTProto mtproto.Options
	// Opened connection limit to the DC.
	MaxOpenConnections int64
	// DC Migration limit.
	MigrationLimit int
}

func (o *Options) setDefaults() {
	o.Device.setDefaults()
	if o.Addr == "" {
		o.Addr = "149.154.167.50:443"
	}
	if o.ID == 0 {
		o.ID = 2
	}
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
	if o.MaxOpenConnections == 0 {
		o.MaxOpenConnections = 2
	}
	if o.MigrationLimit == 0 {
		o.MigrationLimit = 10
	}
}
