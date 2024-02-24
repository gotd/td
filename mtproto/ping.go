package mtproto

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
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
	c.log.Debug("Pong")

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
	// If the client sends these pings once every 60 seconds,
	// for example, it may set disconnect_delay equal to 75 seconds.
	delay := c.pingInterval + c.pingTimeout

	ticker := c.clock.Ticker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "ping loop")
		case <-ticker.C():
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, c.pingTimeout)
				defer cancel()

				return c.pingDelayDisconnect(ctx, int(delay.Seconds()))
			}(); err != nil {
				return errors.Wrap(err, "disconnect (pong missed)")
			}
		}
	}
}
