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
	connectCtx := ctx
	if !c.pfs {
		// Backward-compatible non-PFS behavior: dial timeout limits the whole
		// connect phase (dial + key exchange).
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, c.dialTimeout)
		defer cancel()
	}

	dialCtx := connectCtx
	if c.pfs {
		// Quote (PFS): "The generation of temporary and permanent auth keys can be done in parallel."
		// Link: https://core.telegram.org/api/pfs
		//
		// In PFS mode key generation can take longer, so we only apply dial timeout
		// to socket establishment.
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(ctx, c.dialTimeout)
		defer cancel()
	}

	conn, err := c.dialer(dialCtx)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}
	c.conn = conn
	defer func() {
		if rErr != nil {
			multierr.AppendInto(&rErr, conn.Close())
		}
	}()
	if c.pfs {
		return c.connectPFS(ctx)
	}

	session := c.session()
	if session.Key.Zero() {
		c.log.Info("Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(connectCtx); err != nil {
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

func (c *Conn) connectPFS(ctx context.Context) error {
	if c.permKey.Zero() {
		c.log.Info("Generating new permanent auth key")
		start := c.clock.Now()
		if err := c.createPermAuthKey(ctx); err != nil {
			return errors.Wrap(err, "create permanent auth key")
		}
		c.log.Info("Permanent auth key generated",
			zap.Duration("duration", c.clock.Now().Sub(start)),
		)
	} else {
		// Reuse persisted permanent key to keep existing authorization.
		c.log.Info("Permanent key already exists")
	}

	c.log.Info("Generating new temporary auth key")
	start := c.clock.Now()
	if err := c.createTempAuthKey(ctx); err != nil {
		return errors.Wrap(err, "create temporary auth key")
	}
	c.log.Info("Temporary auth key generated",
		zap.Duration("duration", c.clock.Now().Sub(start)),
	)

	return nil
}

func (c *Conn) runExchange(
	ctx context.Context,
	mode exchange.ExchangeMode,
	expiresIn int,
) (exchange.ClientExchangeResult, error) {
	ex := exchange.NewExchanger(c.conn, c.dcID).
		WithClock(c.clock).
		WithLogger(c.log.Named("exchange")).
		WithTimeout(c.exchangeTimeout).
		WithRand(c.rand)
	if mode == exchange.ExchangeModeTemporary {
		// Temporary mode maps to p_q_inner_data_temp_dc in exchange package.
		ex = ex.WithTempMode(expiresIn)
	}
	return ex.Client(c.rsaPublicKeys).Run(ctx)
}

func (c *Conn) logExchangeInit(ctx context.Context) {
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
}

// createAuthKey generates new authorization key.
func (c *Conn) createAuthKey(ctx context.Context) error {
	// Grab exclusive lock for writing.
	// It prevents message sending during key regeneration if server forgot current auth key.
	c.exchangeLock.Lock()
	defer c.exchangeLock.Unlock()

	c.logExchangeInit(ctx)
	r, err := c.runExchange(ctx, exchange.ExchangeModePermanent, 0)
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

func (c *Conn) createPermAuthKey(ctx context.Context) error {
	c.exchangeLock.Lock()
	defer c.exchangeLock.Unlock()

	c.logExchangeInit(ctx)
	r, err := c.runExchange(ctx, exchange.ExchangeModePermanent, 0)
	if err != nil {
		return err
	}

	c.sessionMux.Lock()
	c.permKey = r.AuthKey
	// Creation timestamp is used by ENCRYPTED_MESSAGE_INVALID recovery policy.
	c.permKeyCreatedAt = c.clock.Now().Unix()
	c.sessionMux.Unlock()

	return nil
}

func (c *Conn) createTempAuthKey(ctx context.Context) error {
	c.exchangeLock.Lock()
	defer c.exchangeLock.Unlock()

	c.logExchangeInit(ctx)
	r, err := c.runExchange(ctx, exchange.ExchangeModeTemporary, c.tempKeyTTL)
	if err != nil {
		return err
	}

	expiresAt := r.ExpiresAt
	if expiresAt == 0 {
		// Defensive fallback if exchange result does not expose expiry.
		expiresAt = c.clock.Now().Unix() + int64(c.tempKeyTTL)
	}

	c.sessionMux.Lock()
	c.authKey = r.AuthKey
	c.sessionID = r.SessionID
	c.salt = r.ServerSalt
	c.tempKeyExpiry = expiresAt
	c.sessionMux.Unlock()

	return nil
}
