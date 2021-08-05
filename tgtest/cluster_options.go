package tgtest

import (
	"crypto/rand"
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// ClusterOptions of Cluster.
type ClusterOptions struct {
	// Web denotes to use websocket listener.
	Web bool
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
	// Ignore opt.Web, it's okay to use zero value.
	if opt.Random == nil {
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	// Ignore opt.Codec, will be handled by transport.Listener.
	// Ignore opt.Config, it's okay to use zero value.
	// Ignore opt.CDNConfig, it's okay to use zero value.
}
