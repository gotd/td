package reliable

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
)

type MTConn interface {
	Run(ctx context.Context, f func(ctx context.Context) error) error
	InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error
}

type MTCreateFunc func(addr string, opts mtproto.Options) MTConn

type Config struct {
	Addr        string
	MTOpts      mtproto.Options
	CreateConn  MTCreateFunc
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
