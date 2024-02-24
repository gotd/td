package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/gotd/td/exchange"
)

// connect establishes connection using configured transport, creating
// new auth key if needed.
func (c *Conn) connect(ctx context.Context) (rErr error) {
	ctx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()

	conn, err := c.dialer(ctx)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}
	c.conn = conn
	defer func() {
		if rErr != nil {
			multierr.AppendInto(&rErr, conn.Close())
		}
	}()

	session := c.session()
	if session.Key.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(ctx); err != nil {
			return errors.Wrap(err, "create auth key")
		}

		c.log.Info("Auth key generated",
			zap.Duration("duration", c.clock.Now().Sub(start)),
		)
		return nil
	}

	c.log.Info("Key already exists")
	if session.ID == 0 {
		// NB: Telegram can return 404 error if session id is zero.
		//
		// See https://github.com/gotd/td/issues/107.
		c.log.Debug("Generating new session id")
		if err := c.newSessionID(); err != nil {
			return err
		}
	}

	return nil
}

// createAuthKey generates new authorization key.
func (c *Conn) createAuthKey(ctx context.Context) error {
	// Grab exclusive lock for writing.
	// It prevents message sending during key regeneration if server forgot current auth key.
	c.exchangeLock.Lock()
	defer c.exchangeLock.Unlock()

	if ce := c.log.Check(zap.DebugLevel, "Initializing new key exchange"); ce != nil {
		// Useful for debugging i/o timeout errors on tcp reads or writes.
		fields := []zap.Field{
			zap.Duration("timeout", c.exchangeTimeout),
		}
		if deadline, ok := ctx.Deadline(); ok {
			fields = append(fields, zap.Time("context_deadline", deadline))
		}
		ce.Write(fields...)
	}

	r, err := exchange.NewExchanger(c.conn, c.dcID).
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
