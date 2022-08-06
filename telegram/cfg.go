package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// Available MTProto default server addresses.
//
// See https://my.telegram.org/apps.
const (
	AddrProduction = "149.154.167.50:443"
	AddrTest       = "149.154.167.40:443"
)

// Test-only credentials. Can be used with AddrTest and TestAuth to
// test authentication.
//
// Reference:
//   - https://github.com/telegramdesktop/tdesktop/blob/5f665b8ecb48802cd13cfb48ec834b946459274a/docs/api_credentials.md
const (
	TestAppID   = constant.TestAppID
	TestAppHash = constant.TestAppHash
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
