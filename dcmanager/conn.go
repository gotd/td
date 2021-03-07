package dcmanager

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
)

type Conn interface {
	Run(ctx context.Context, f func(ctx context.Context) error) error
	InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error
}

type CreateConnFunc func(addr string, opts mtproto.Options) Conn
