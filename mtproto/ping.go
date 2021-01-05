package mtproto

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
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

	if err := c.writeServiceMessage(ctx, &mt.PingRequest{PingID: pingID}); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	return c.waitPong(ctx, pingID)
}

func (c *Conn) handlePong(b *bin.Buffer) error {
	var pong mt.Pong
	if err := pong.Decode(b); err != nil {
		return xerrors.Errorf("decode: %x", err)
	}
	c.log.Debug("Pong")

	c.pingMux.Lock()
	f, ok := c.ping[pong.PingID]
	c.pingMux.Unlock()
	if ok {
		f()
	}
	return nil
}

func (c *Conn) pingDelayDisconnect(ctx context.Context, delay int) error {
	// Generating random id.
	// Probably we should check for collisions here.
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	if err := c.writeServiceMessage(ctx, &mt.PingDelayDisconnectRequest{
		PingID:          pingID,
		DisconnectDelay: delay,
	}); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	return c.waitPong(ctx, pingID)
}

func (c *Conn) waitPong(ctx context.Context, pingID int64) error {
	// Setting ping callback before write.
	result := make(chan struct{})
	c.pingMux.Lock()
	c.ping[pingID] = func() {
		close(result)
	}
	c.pingMux.Unlock()

	defer func() {
		c.pingMux.Lock()
		delete(c.ping, pingID)
		c.pingMux.Unlock()
	}()

	// Waiting for result.
	select {
	case <-result:
		// Received pong with pingID.
		return nil
	case <-ctx.Done():
		// Something gone really bad.
		return ctx.Err()
	}
}

func (c *Conn) pingLoop(ctx context.Context) error {
	const (
		timeout   = time.Second * 15
		frequency = time.Minute
		// If the client sends these pings once every 60 seconds,
		// for example, it may set disconnect_delay equal to 75 seconds.
		disconnectDelay = 75 // in seconds
	)

	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return xerrors.Errorf("ping loop: %w", ctx.Err())
		case <-ticker.C:
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				return c.pingDelayDisconnect(ctx, disconnectDelay)
			}(); err != nil {
				return xerrors.Errorf("disconnect (pong missed): %w", err)
			}
		}
	}
}
