package reliable

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
)

// MTConn represents MTProto connection.
type MTConn interface {
	Run(ctx context.Context, f func(ctx context.Context) error) error
	InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error
}

// Config of reliable conn.
type Config struct {
	// Telegram's server address.
	Addr string
	// MTProto options.
	MTOpts mtproto.Options
	// Function which creates uninitialized MTProto connection.
	CreateConn func(addr string, opts mtproto.Options) MTConn
	// Callback to actions which must be performed before use the connection.
	// Typically 'tg.InitConnectionRequest'.
	OnConnected func(MTConn) error
}

func (cfg *Config) setDefaults() {
	if cfg.CreateConn == nil {
		cfg.CreateConn = func(addr string, opts mtproto.Options) MTConn {
			return mtproto.New(addr, opts)
		}
	}

	if cfg.OnConnected == nil {
		cfg.OnConnected = func(m MTConn) error { return nil }
	}
}
