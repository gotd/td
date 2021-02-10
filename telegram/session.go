package telegram

import (
	"sync/atomic"

	"github.com/gotd/td/tg"
)

type atomicConfig struct {
	atomic.Value
}

func (c *atomicConfig) Load() tg.Config {
	return c.Value.Load().(tg.Config)
}

func (c *atomicConfig) Store(cfg tg.Config) {
	c.Value.Store(cfg)
}
