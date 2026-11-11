package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
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

// Ping cadence of Telegram Desktop (mtproto/session_private.cpp): the next
// ping is due kPingSendAfter after the last one and is forced at
// kPingSendAfterForce; ping_delay_disconnect always asks for 60 seconds.
const (
	pingSendAfter       = 30 * time.Second
	pingSendAfterForce  = 45 * time.Second
	pingDelayDisconnect = 60
)

// nextPing returns the time until the next ping and its disconnect_delay.
//
// With the default (zero) pingInterval, the interval is random in
// [pingSendAfter, pingSendAfterForce]: Telegram Desktop sends a ping at the
// first send after pingSendAfter or by timer at pingSendAfterForce, which
// gives 30–45 s on the wire.
func (c *Conn) nextPing() (time.Duration, int) {
	if c.pingInterval > 0 {
		// If the client sends these pings once every 60 seconds,
		// for example, it may set disconnect_delay equal to 75 seconds.
		return c.pingInterval, int((c.pingInterval + c.pingTimeout).Seconds())
	}
	jitter := pingSendAfterForce - pingSendAfter
	n, err := crypto.RandInt64n(c.rand, int64(jitter)+1)
	if err != nil {
		n = int64(jitter / 2)
	}
	return pingSendAfter + time.Duration(n), pingDelayDisconnect
}

func (c *Conn) pingLoop(ctx context.Context) error {
	logger := c.log.Named("ping")
	interval, delay := c.nextPing()
	timer := c.clock.Timer(interval)
	defer func() { clock.StopTimer(timer) }()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "ping loop")
		case <-timer.C():
			start := c.clock.Now()
			logger.Debug(ctx, "Sending ping", log.Duration("timeout", c.pingTimeout))
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, c.pingTimeout)
				defer cancel()

				return c.pingDelayDisconnect(ctx, delay)
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
			elapsed := c.clock.Now().Sub(start)
			logger.Debug(ctx, "Ping acknowledged", log.Duration("elapsed", elapsed))

			// The next ping is due interval after this one was sent.
			interval, delay = c.nextPing()
			clock.StopTimer(timer)
			timer = c.clock.Timer(max(interval-elapsed, 0))
		}
	}
}
