package telegram

import (
	"context"
)

// Ping sends low level ping request to Telegram server.
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}
