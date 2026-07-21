package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"
	"go.uber.org/multierr"

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

	// Declared right after connectCtx's own "defer cancel()" above (not
	// before it): defers are LIFO, so this runs BEFORE that cancel() fires,
	// while still running AFTER every other cleanup defer below (closing
	// conn, closing connectDone) since those are declared later and so run
	// earlier. Getting this backwards silently breaks the check: cancel()
	// fires on every return path including plain success, so testing
	// connectCtx.Err() from a defer declared before it would see a non-nil
	// Err() unconditionally and rewrite every genuine error (bad nonce,
	// malformed response, ...) into context.Canceled.
	//
	// Normalizes a non-nil result to the caller's ctx.Err() if the caller
	// cancelled ctx, or to connectCtx.Err() if connectCtx expired on its own
	// (dialTimeout firing mid-exchange while ctx is still live — the watcher
	// below observes connectCtx, not ctx, precisely to force-close on that
	// case too). Otherwise the forced Close() from the watcher goroutine
	// below surfaces as a raw transport error (e.g. "use of closed network
	// connection"), which breaks callers that use errors.Is(err,
	// context.Canceled / context.DeadlineExceeded) to distinguish
	// cancellation/timeout from a genuine transport failure.
	// Normalization is lossy by design (see above), which leaves nothing to go
	// on when debugging a dial failure that merely coincided with
	// cancellation: the transport error naming the actual cause is gone by the
	// time the caller sees the result. Log it before overwriting rErr — debug
	// level, so it costs nothing when disabled.
	defer func() {
		if rErr == nil {
			return
		}
		if ctx.Err() != nil {
			c.log.Debug(ctx, "Connect failed, reporting context error instead",
				log.Error(rErr),
			)
			rErr = errors.Wrap(ctx.Err(), "connect")
			return
		}
		if errors.Is(connectCtx.Err(), context.DeadlineExceeded) {
			c.log.Debug(ctx, "Connect failed, reporting dial timeout instead",
				log.Error(rErr),
			)
			rErr = errors.Wrap(connectCtx.Err(), "connect")
		}
	}()

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
	// Watchdog must not fire on a freshly established connection before
	// anything has been read yet.
	c.lastRecv.Store(c.clock.Now().UnixNano())

	// The watcher below and the close-on-error defer can both decide to
	// close conn: the watcher fires on connectCtx.Done(), the defer fires
	// whenever connect() returns a non-nil error, and neither implies the
	// other (e.g. connectCtx expiring and createAuthKey failing can race).
	// transport.Conn does not promise concurrent-Close safety, so route both
	// — and every other closer for the lifetime of this Conn (handleClose,
	// idleWatchdog; see mtproto/conn.go) — through Conn.close, which guards
	// c.conn.Close() with one Conn-scoped sync.Once.
	defer func() {
		if rErr != nil {
			multierr.AppendInto(&rErr, c.close())
		}
	}()

	// Key exchange runs before Conn.Run starts its goroutine group, so
	// handleClose — which is what normally closes the socket on cancellation —
	// does not exist yet. Without this watcher a parked exchange read ignores
	// ctx entirely, and callers that cancel and wait (pool.DC.Close ->
	// Supervisor.Wait) hang forever.
	//
	// Watches connectCtx, not ctx: connectCtx is what actually drives
	// createAuthKey below. In PFS mode connectCtx is ctx, so this is
	// identical to watching ctx. In non-PFS mode connectCtx additionally
	// expires on dialTimeout, which must also force-close the socket — a
	// net.Conn wrapper that ignores context deadlines (e.g. websocket,
	// mtproxy) would otherwise stay parked past the dial timeout with
	// nothing left to interrupt it.
	connectDone := make(chan struct{})
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-connectCtx.Done():
			if err := c.close(); err != nil {
				c.log.Debug(ctx, "Failed to close connection on cancel", log.Error(err))
			}
		case <-connectDone:
		}
	}()
	// In non-PFS mode connectCtx has its own "defer cancel()" above, which
	// fires on every return path including plain success — not just genuine
	// expiry — and closes connectCtx.Done() at (as far as another goroutine
	// can observe) the same instant as connectDone below. Two simultaneously
	// ready select cases are chosen at random, so without this wait the
	// watcher could still race-close a connection that connect() already
	// returned successfully. Waiting for watcherDone here, before that
	// cancel() defer gets a chance to run, guarantees the watcher commits to
	// a branch while connectCtx.Done() can only be ready for a real reason
	// (genuine dial-timeout expiry or caller cancellation during the
	// exchange), not as a side effect of connect() itself returning.
	//
	// Trade-off, kept deliberately: this wait still bounds how long the
	// connect-phase watcher goroutine above can outlive connect() itself.
	// Its Close() calls now route through Conn.close (mtproto/conn.go),
	// which shares one Conn-scoped sync.Once with handleClose and
	// idleWatchdog, so a connect-phase Close() still in flight when Run
	// starts its goroutine group no longer races a second, unguarded
	// Close() against the same conn — that hazard originally motivated this
	// wait, and is now closed at the source by the shared guard. What
	// remains is goroutine hygiene: without this wait, Run could start its
	// group while this watcher goroutine is still parked on
	// connectCtx.Done(), leaving a stray goroutine whose lifetime crosses
	// the connect()/Run() boundary. The cost: Conn.close's Close() call can
	// itself block, since transport/connection.go calls Close() on an
	// arbitrary net.Conn, and a wrapper (websocket, mtproxy, SOCKS) with a
	// blocking Close() would stall this wait — and therefore connect()'s
	// return — even on the success path, in the nanosecond window where
	// connectCtx.Done() becomes ready for a genuine reason at essentially
	// the same instant the exchange itself completes. That window is
	// narrow and the goroutine-lifetime property this wait buys is real,
	// so keep it; removing it is a separate decision, not implied by the
	// guard becoming shared.
	defer func() {
		close(connectDone)
		<-watcherDone
	}()

	if c.pfs {
		return c.connectPFS(ctx)
	}

	session := c.session()
	if session.Key.Zero() {
		c.log.Info(ctx, "Generating new auth key")
		start := c.clock.Now()
		if err := c.createAuthKey(connectCtx); err != nil {
			return errors.Wrap(err, "create auth key")
		}

		c.log.Info(ctx, "Auth key generated",
			log.Duration("duration", c.clock.Now().Sub(start)),
		)
		return nil
	}

	c.log.Info(ctx, "Key already exists")
	if session.ID == 0 {
		// NB: Telegram can return 404 error if session id is zero.
		//
		// See https://github.com/gotd/td/issues/107.
		c.log.Debug(ctx, "Generating new session id")
		if err := c.newSessionID(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Conn) connectPFS(ctx context.Context) error {
	if c.permKey.Zero() {
		c.log.Info(ctx, "Generating new permanent auth key")
		start := c.clock.Now()
		if err := c.createPermAuthKey(ctx); err != nil {
			return errors.Wrap(err, "create permanent auth key")
		}
		c.log.Info(ctx, "Permanent auth key generated",
			log.Duration("duration", c.clock.Now().Sub(start)),
		)
	} else {
		// Reuse persisted permanent key to keep existing authorization.
		c.log.Info(ctx, "Permanent key already exists")
	}

	c.log.Info(ctx, "Generating new temporary auth key")
	start := c.clock.Now()
	if err := c.createTempAuthKey(ctx); err != nil {
		return errors.Wrap(err, "create temporary auth key")
	}
	c.log.Info(ctx, "Temporary auth key generated",
		log.Duration("duration", c.clock.Now().Sub(start)),
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
		WithLogger(c.log.Named("exchange").Logger()).
		WithTimeout(c.exchangeTimeout).
		WithRand(c.rand)
	if mode == exchange.ExchangeModeTemporary {
		// Temporary mode maps to p_q_inner_data_temp_dc in exchange package.
		ex = ex.WithTempMode(expiresIn)
	}
	return ex.Client(c.rsaPublicKeys).Run(ctx)
}

func (c *Conn) logExchangeInit(ctx context.Context) {
	if !c.log.Enabled(ctx, log.LevelDebug) {
		return
	}
	// Useful for debugging i/o timeout errors on tcp reads or writes.
	attrs := []log.Attr{
		log.Duration("timeout", c.exchangeTimeout),
	}
	if deadline, ok := ctx.Deadline(); ok {
		attrs = append(attrs, log.Time("context_deadline", deadline))
	}
	c.log.Debug(ctx, "Initializing new key exchange", attrs...)
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
