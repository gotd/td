package telegram

import "github.com/gotd/td/internal/crypto"

func (c *Client) RandInt64() (int64, error) {
	return crypto.RandInt64(c.rand)
}
