package mtproto

import (
	"context"

	"github.com/gotd/td/internal/exchange"
)

// createAuthKey generates new authorization key.
func (c *Conn) createAuthKey(ctx context.Context) error {
	r, err := exchange.NewExchanger(c.conn).
		WithClock(c.clock).
		WithLogger(c.log.Named("exchange")).
		WithTimeout(c.exchangeTimeout).
		WithRand(c.rand).
		Client(c.rsaPublicKeys).Run(ctx)
	if err != nil {
		return err
	}

	c.sessionMux.Lock()
	c.authKey = r.AuthKey
	c.sessionID = r.SessionID
	c.salt = r.ServerSalt
	c.sessionMux.Unlock()

	return nil
}
