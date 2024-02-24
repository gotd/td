package rpc

import (
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/clock"
)

// Options of rpc engine.
type Options struct {
	RetryInterval time.Duration
	MaxRetries    int
	Logger        *zap.Logger
	Clock         clock.Clock
	DropHandler   DropHandler
}

func (cfg *Options) setDefaults() {
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = time.Second * 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.System
	}
	if cfg.DropHandler == nil {
		cfg.DropHandler = NopDrop
	}
}
