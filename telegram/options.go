package telegram

import (
	"crypto/rand"
	"io"

	"github.com/gotd/td/mtproto"
	"go.uber.org/zap"
)

// Options of telegram client
type Options struct {
	UpdateHandler UpdateHandler
	Random        io.Reader
	MTProto       mtproto.Options
	Logger        *zap.Logger
}

func (opts *Options) setDefaults() {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	if opts.Random == nil {
		opts.Random = rand.Reader
	}
}
