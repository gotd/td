package telegram

import (
	"github.com/nnqq/td/telegram/auth"
)

// Auth returns auth client.
func (c *Client) Auth() *auth.Client {
	return auth.NewClient(
		c.tg, c.rand, c.appID, c.appHash,
	)
}

func unauthorized(err error) bool {
	return auth.IsKeyUnregistered(err)
}
