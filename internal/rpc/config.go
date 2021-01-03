package rpc

import (
	"time"

	"go.uber.org/zap"
)

// Clock abstracts temporal effects.
type Clock interface {
	After(d time.Duration) <-chan time.Time
}

type systemClock struct{}

func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Config of rpc engine.
type Config struct {
	RetryInterval time.Duration
	MaxRetries    int
	Logger        *zap.Logger
	Clock         Clock
}

func (cfg *Config) setDefaults() {
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
		cfg.Clock = systemClock{}
	}
}
