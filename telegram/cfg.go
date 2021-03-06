package telegram

import (
	"context"

	"go.uber.org/zap"
)

func (c *Client) fetchConfig(ctx context.Context) {
	cfg, err := c.tg.HelpGetConfig(ctx)
	if err != nil {
		c.log.Warn("Got error on config update", zap.Error(err))
		return
	}

	c.cfg.Store(*cfg)
}
