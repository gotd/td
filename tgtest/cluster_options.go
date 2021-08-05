package tgtest

import (
	"context"
	"crypto/rand"
	"io"
	"net"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// ListenFunc is a simple alias for listener factory.
type ListenFunc = func(ctx context.Context, dc int) (net.Listener, error)

// ClusterOptions of Cluster.
type ClusterOptions struct {
	// Listen creates new net.Listener for DC.
	// Defaults to net.ListenConfig with random port.
	Listen ListenFunc
	// Random is random source. Used to generate RSA keys.
	// Defaults to rand.Reader.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// Codec constructor.
	// Defaults to nil (underlying transport server detects protocol automatically).
	Codec func() transport.Codec
	// Config is a initial cluster config.
	Config tg.Config
	// CDNConfig is a initial cluster CDN config.
	CDNConfig tg.CDNConfig
}

func (opt *ClusterOptions) setDefaults() {
	if opt.Listen == nil {
		opt.Listen = func(ctx context.Context, dc int) (net.Listener, error) {
			conf := net.ListenConfig{}
			l, err := conf.Listen(ctx, "tcp4", "127.0.0.1:0")
			if err != nil {
				return nil, err
			}
			return l, nil
		}
	}
	if opt.Random == nil {
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}

	// Ignore opt.Codec, will be handled by transport.NewCustomServer.
	// Ignore opt.Config, it's okay to use zero value.
	// Ignore opt.CDNConfig, it's okay to use zero value.
}
