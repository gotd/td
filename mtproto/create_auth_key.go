package mtproto

import (
	"context"
	"sync/atomic"

	"github.com/gotd/td/internal/exchange"
)

// createAuthKey generates new authorization key.
func (c *Conn) createAuthKey(ctx context.Context) error {
	r, err := exchange.NewExchanger(c.conn).
		WithClock(c.clock).
		WithLogger(c.log.Named("exchange")).
		WithRand(c.rand).
		Client(c.rsaPublicKeys).Run(ctx)
	if err != nil {
		return err
	}

	c.authKey = r.AuthKey
	atomic.StoreInt64(&c.session, r.SessionID)
	atomic.StoreInt64(&c.salt, r.ServerSalt)

	return nil
}
