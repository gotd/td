package telegram

import (
	"context"
	"sync/atomic"

	"github.com/gotd/td/telegram/internal/exchange"
)

// createAuthKey generates new authorization key.
func (c *Client) createAuthKey(ctx context.Context) error {
	cfg := exchange.NewConfig(c.clock, c.rand, c.conn, c.log.Named("exchange"))
	r, err := exchange.NewClientExchange(cfg, c.rsaPublicKeys...).Run(ctx)
	if err != nil {
		return err
	}

	c.authKey = r.AuthKey
	atomic.StoreInt64(&c.session, r.SessionID)
	atomic.StoreInt64(&c.salt, r.ServerSalt)

	return nil
}
