package telegram

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/log"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// runUntilRestart runs the primary connection until restart. It stores the
// DC of the connection in dc, so that its state is reported for the right DC
// even after migration changed the session.
func (c *Client) runUntilRestart(ctx context.Context, dc *int) error {
	c.connMux.Lock()
	conn := c.conn
	*dc = c.connDC
	c.connMux.Unlock()

	c.notifyConnectionState(ConnectionStateConnecting, *dc)

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(conn.Run)

	// Report readiness of this primary connection only.
	//
	// We must not emit this from the shared OnSession handler, because it is
	// also used by pool/sub connections (including same-DC ones), which would
	// produce spurious ready events. Waiting on the connection's own Ready
	// channel scopes the event strictly to the primary connection.
	if c.onConnectionState != nil || c.onConnectionStateEvent != nil {
		connDC := *dc
		if ready, ok := conn.(interface{ Ready() <-chan struct{} }); ok {
			g.Go(func(ctx context.Context) error {
				if !waitReady(ctx, ready.Ready()) {
					return ctx.Err()
				}
				c.notifyConnectionState(ConnectionStateReady, connDC)
				return nil
			})
		}
	}

	// If we don't need updates, so there is no reason to subscribe for it.
	if !c.noUpdatesMode {
		g.Go(func(ctx context.Context) error {
			// Call method which requires authorization, to subscribe for updates.
			// See https://core.telegram.org/api/updates#subscribing-to-updates.
			self, err := c.Self(ctx)
			if err != nil {
				// Ignore unauthorized errors.
				if !auth.IsUnauthorized(err) {
					c.log.Warn(ctx, "Got error on self", log.Error(err))
				}
				if IsAuthLost(err) {
					// OnAuthLost runs on its own goroutine: wait for it to
					// know whether it stops Run.
					if err := c.waitAuthLost(ctx); err != nil {
						return err
					}
				}
				if c.auth.stopped.Load() {
					// OnAuthLost already reported this error and stops Run.
					return nil
				}
				if h := c.onSelfError; h != nil {
					// Help with https://github.com/gotd/td/issues/1458.
					if err := h(ctx, err); err != nil {
						return errors.Wrap(err, "onSelfError")
					}
				}
				return nil
			}

			c.markAuthorized()
			if h := c.onSelfSuccess; h != nil {
				h(self)
			}

			c.log.Info(ctx, "Got self", log.String("username", self.Username))
			return nil
		})
	}

	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.restart:
			c.log.Debug(ctx, "Restart triggered")
			// Should call cancel() to cancel group.
			g.Cancel()

			return nil
		}
	})

	return g.Wait()
}

// waitReady blocks until ready is closed or ctx is done and reports whether
// ready was closed.
//
// Readiness wins when both happen: a request may complete on the fresh
// connection and cancel ctx before this goroutine is scheduled, and a plain
// select would then drop the ready event at random.
func waitReady(ctx context.Context, ready <-chan struct{}) bool {
	select {
	case <-ready:
		return true
	case <-ctx.Done():
		select {
		case <-ready:
			return true
		default:
			return false
		}
	}
}

func (c *Client) isPermanentError(err error) bool {
	// See https://github.com/gotd/td/issues/1458.
	if errors.Is(err, exchange.ErrKeyFingerprintNotFound) {
		return true
	}
	if tgerr.Is(err, "AUTH_KEY_UNREGISTERED", "SESSION_EXPIRED", "AUTH_KEY_DUPLICATED") {
		return true
	}
	if auth.IsUnauthorized(err) {
		return true
	}
	return false
}

func (c *Client) reconnectUntilClosed(ctx context.Context) error {
	// Note that we currently have no timeout on connection, so this is
	// potentially eternal.
	b := tdsync.SyncBackoff(backoff.WithContext(c.newConnBackoff(), ctx))
	c.connBackoff.Store(&b)

	// DC of the last run primary connection.
	var dc int
	return backoff.RetryNotify(func() error {
		if err := c.runUntilRestart(ctx, &dc); err != nil {
			if c.isPermanentError(err) {
				c.log.Warn(ctx, "Permanent connection error, will not reconnect", log.Error(err))
				return backoff.Permanent(err)
			}
			return err
		}

		return nil
	}, b, func(err error, timeout time.Duration) {
		c.log.Info(context.Background(), "Restarting connection", log.Error(err), log.Duration("backoff", timeout))
		c.notifyConnectionState(ConnectionStateDisconnected, dc)

		c.connMux.Lock()
		// Some PFS errors require dropping persisted keys before recreating conn.
		c.handlePrimaryConnDead(err)
		c.replaceConn(c.createPrimaryConn(nil))
		c.connMux.Unlock()
		c.log.Debug(context.Background(), "Primary connection replaced after restart")
	})
}

func (c *Client) onReady() {
	c.log.Debug(context.Background(), "Ready")
	c.ready.Signal()

	if b := c.connBackoff.Load(); b != nil {
		// Reconnect faster next time.
		(*b).Reset()
	}
}

func (c *Client) resetReady() {
	c.ready.Reset()
}

// Run starts client session and blocks until connection close.
// The f callback is called on successful session initialization and Run
// will return on f() result.
//
// Context of callback will be canceled if fatal error is detected.
// The ctx is used for background operations like updates handling or pools.
//
// See `examples/bg-run` and `contrib/gb` package for classic approach without
// explicit callback, with Connect and defer close().
func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) (err error) {
	if c.ctx != nil {
		select {
		case <-c.ctx.Done():
			return errors.Wrap(c.ctx.Err(), "client already closed")
		default:
		}
	}

	// Setting up client context for background operations like updates
	// handling or pool creation.
	c.ctx, c.cancel = context.WithCancel(ctx)

	c.log.Info(ctx, "Starting")
	defer c.log.Info(context.Background(), "Closed")
	// Wait for a running OnAuthLost call after c.ctx (passed to it) is
	// canceled and pools are closed below, so that it doesn't outlive Run.
	// If it asked to stop Run, report that even if f returned first.
	defer func() {
		_ = c.waitAuthLost(context.Background())
		select {
		case lost := <-c.authLost:
			var already *AuthLostError
			switch {
			case err == nil:
				err = lost
			case !errors.As(err, &already):
				err = multierr.Append(lost, err)
			}
		default:
		}
	}()
	defer func() {
		subConns := c.subConns.drain()
		subConns = append(subConns, c.filePools.drain()...)
		c.fileInvokersMux.Lock()
		c.fileInvokers = map[int]tg.Invoker{}
		c.fileInvokersMux.Unlock()

		cdnConns := c.cdnPools.drain()

		// Close outside locks to avoid lock inversion with pool callbacks.
		for _, conn := range subConns {
			if closeErr := conn.Close(); !errors.Is(closeErr, context.Canceled) {
				multierr.AppendInto(&err, closeErr)
			}
		}
		for _, conn := range cdnConns {
			if closeErr := conn.Close(); !errors.Is(closeErr, context.Canceled) {
				multierr.AppendInto(&err, closeErr)
			}
		}
	}()
	// Cancel client before deferred cleanup snapshot to block concurrent pool
	// creations while we are draining cached connections.
	defer c.cancel()

	c.resetReady()
	if err := c.restoreConnection(ctx); err != nil {
		return err
	}

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.reconnectUntilClosed)
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-c.authLost:
			return err
		}
	})
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			c.cancel()
			return ctx.Err()
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	})
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.ready.Ready():
			if err := f(ctx); err != nil {
				return errors.Wrap(err, "callback")
			}
			// Should call cancel() to cancel ctx.
			// This will terminate c.conn.Run().
			c.log.Debug(ctx, "Callback returned, stopping")
			g.Cancel()
			return nil
		}
	})
	if err := g.Wait(); !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
