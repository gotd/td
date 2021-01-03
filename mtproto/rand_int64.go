package mtproto

import "github.com/gotd/td/internal/crypto"

// RandInt64 returns new random int64 from random source.
//
// Useful helper for places in API where random int is required.
func (c *Conn) RandInt64() (int64, error) {
	return crypto.RandInt64(c.rand)
}
