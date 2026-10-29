package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/transport"
)

// Ping sends ping request to server and waits until pong is received or
// context is canceled.
func (c *Conn) Ping(ctx context.Context) error {
	// Generating random id.
	// Probably we should check for collisions here.
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	pong := c.pong(pingID)
	defer c.removePong(pingID)

	if err := c.writeServiceMessage(ctx, &mt.PingRequest{PingID: pingID}); err != nil {
		return errors.Wrap(err, "write")
	}

	select {
	case <-pong:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Conn) handlePong(b *bin.Buffer) error {
	var pong mt.Pong
	if err := pong.Decode(b); err != nil {
		return errors.Errorf("decode: %x", err)
	}
	// Any pong, not only the one this connection is currently waiting for.
	// Progress-aware mode samples this counter from the ping loop.
	c.pongSeq.Inc()
	c.log.Debug(context.Background(), "Pong")

	c.pingMux.Lock()
	ch, ok := c.ping[pong.PingID]
	if ok {
		close(ch)
		delete(c.ping, pong.PingID)
	}
	c.pingMux.Unlock()

	return nil
}

func (c *Conn) pingDelayDisconnect(ctx context.Context, delay int) error {
	// Generating random id.
	// Probably we should check for collisions here.
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	pong := c.pong(pingID)
	defer c.removePong(pingID)

	if err := c.writeServiceMessage(ctx, &mt.PingDelayDisconnectRequest{
		PingID:          pingID,
		DisconnectDelay: delay,
	}); err != nil {
		return errors.Wrap(err, "write")
	}

	select {
	case <-pong:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Conn) pong(pingID int64) chan struct{} {
	ch := make(chan struct{})
	c.pingMux.Lock()
	c.ping[pingID] = ch
	c.pingMux.Unlock()
	return ch
}

func (c *Conn) removePong(pingID int64) {
	c.pingMux.Lock()
	delete(c.ping, pingID)
	c.pingMux.Unlock()
}

func (c *Conn) pingLoop(ctx context.Context) error {
	if c.pingProgressAware {
		return c.pingLoopProgress(ctx)
	}
	return c.pingLoopLegacy(ctx)
}

// progress returns a monotonically increasing counter of inbound activity.
func (c *Conn) progress() uint64 {
	p := c.recvMessages.Load()
	if rc, ok := c.conn.(transport.ReadCounter); ok {
		p += rc.BytesRead()
	}
	return p
}

func (c *Conn) pingLoopLegacy(ctx context.Context) error {
	// If the client sends these pings once every 60 seconds,
	// for example, it may set disconnect_delay equal to 75 seconds.
	delay := c.pingInterval + c.pingTimeout

	logger := c.log.Named("ping")
	ticker := c.clock.Ticker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "ping loop")
		case <-ticker.C():
			start := c.clock.Now()
			logger.Debug(ctx, "Sending ping", log.Duration("timeout", c.pingTimeout))
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, c.pingTimeout)
				defer cancel()

				return c.pingDelayDisconnect(ctx, int(delay.Seconds()))
			}(); err != nil {
				// A missed pong means the connection is half-open: the server
				// is not answering our pings even though the read loop may
				// still be delivering updates. This should trigger a reconnect.
				logger.Warn(ctx, "Ping failed, connection considered dead",
					log.Error(err),
					log.Duration("elapsed", c.clock.Now().Sub(start)),
				)
				return errors.Wrap(err, "disconnect (pong missed)")
			}
			logger.Debug(ctx, "Ping acknowledged", log.Duration("elapsed", c.clock.Now().Sub(start)))
		}
	}
}

func (c *Conn) pingLoopProgress(ctx context.Context) error {
	logger := c.log.Named("ping")
	delay := int((c.pingInterval + c.pingTimeout).Seconds())

	send := c.clock.Ticker(c.pingInterval)
	defer send.Stop()
	check := c.clock.Ticker(pingCheckInterval(c.pingTimeout))
	defer check.Stop()

	t := newPingTracker(c.clock.Now(), c.pingTimeout, c.pingMaxWait, c.progress(), c.pongSeq.Load())
	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "ping loop")
		case <-send.C():
			now := c.clock.Now()
			t.observe(now, c.progress(), c.pongSeq.Load())
			if err := c.sendPing(ctx, delay); err != nil {
				logger.Warn(ctx, "Ping failed, connection considered dead",
					log.Error(err),
					log.String("reason", "write"),
				)
				return errors.Wrap(err, "disconnect (ping write)")
			}
			t.onSend(now)
		case <-check.C():
			now := c.clock.Now()
			t.observe(now, c.progress(), c.pongSeq.Load())
			if reason, dead := t.dead(now); dead {
				logger.Warn(ctx, "Ping failed, connection considered dead",
					log.String("reason", reason),
					log.Duration("pending_for", now.Sub(t.pendingSince)),
					log.Duration("since_progress", now.Sub(t.lastAlive)),
				)
				return errors.Errorf("disconnect (pong missed): %s", reason)
			}
		}
	}
}

// sendPing writes ping_delay_disconnect without waiting for the pong.
func (c *Conn) sendPing(ctx context.Context, delay int) error {
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pingTimeout)
	defer cancel()
	return c.writeServiceMessage(ctx, &mt.PingDelayDisconnectRequest{
		PingID:          pingID,
		DisconnectDelay: delay,
	})
}
