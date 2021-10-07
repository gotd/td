package telegram

import "github.com/nnqq/td/internal/crypto"

// RandInt64 returns new random int64 from random source.
//
// Useful helper for places in API where random int is required.
func (c *Client) RandInt64() (int64, error) {
	return crypto.RandInt64(c.rand)
}
