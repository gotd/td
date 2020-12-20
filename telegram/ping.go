package telegram

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
)

type pingMessage struct {
	id int64
}

func (p pingMessage) Encode(b *bin.Buffer) error {
	b.PutID(0x7abe77ec)
	b.PutLong(p.id)
	return nil
}

// Ping sends ping request to server and waits until pong is received or
// context is canceled.
func (c *Client) Ping(ctx context.Context) error {
	// Generating random id.
	// Probably we should check for collisions here.
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	if err := c.writeServiceMessage(ctx, pingMessage{id: pingID}); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	return c.waitPong(ctx, pingID)
}

func (c *Client) handlePong(b *bin.Buffer) error {
	var pong mt.Pong
	if err := pong.Decode(b); err != nil {
		return xerrors.Errorf("decode: %x", err)
	}
	c.log.Info("Pong")

	c.pingMux.Lock()
	f, ok := c.ping[pong.PingID]
	c.pingMux.Unlock()
	if ok {
		f()
	}
	return nil
}

type pingDelayDisconnectMessage struct {
	id    int64
	delay int // in seconds
}

func (p pingDelayDisconnectMessage) Encode(b *bin.Buffer) error {
	b.PutID(0xf3427b8c)
	b.PutLong(p.id)
	b.PutInt(int(p.delay))
	return nil
}

func (c *Client) pingDelayDisconnect(ctx context.Context, delay int) error {
	// Generating random id.
	// Probably we should check for collisions here.
	pingID, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	if err := c.writeServiceMessage(ctx, pingDelayDisconnectMessage{
		id:    pingID,
		delay: delay,
	}); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	return c.waitPong(ctx, pingID)
}

func (c *Client) waitPong(ctx context.Context, pingID int64) error {
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

func (c *Client) pingLoop(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	log := c.log.Named("pinger")

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
			return
		case <-ticker.C:
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				return c.pingDelayDisconnect(ctx, disconnectDelay)
			}(); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}

				log.Warn("ping error", zap.Error(err))
				ticker.Stop()

				if err := c.reconnect(); err != nil {
					// TODO(ccln): what next???
					log.Error("reconnect", zap.Error(err))
				}

				ticker.Reset(frequency)
			}
		}
	}
}
