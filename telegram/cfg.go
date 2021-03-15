package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

// Config returns current config.
func (c *Client) Config() tg.Config {
	return c.cfg.Load()
}

func (c *Client) fetchConfig(ctx context.Context) {
	cfg, err := c.tg.HelpGetConfig(ctx)
	if err != nil {
		c.log.Warn("Got error on config update", zap.Error(err))
		return
	}

	c.cfg.Store(*cfg)
}
