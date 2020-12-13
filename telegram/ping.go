package telegram

import (
	"context"
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

	// Setting ping callback before write.
	result := make(chan struct{})
	c.pingMux.Lock()
	c.ping[pingID] = func() {
		close(result)
	}
	c.pingMux.Unlock()

	if err := c.write(c.newMessageID(), c.seqNo(), pingMessage{id: pingID}); err != nil {
		return xerrors.Errorf("failed to write: %w", err)
	}

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

func (c *Client) handlePong(b *bin.Buffer) error {
	var pong mt.Pong
	if err := pong.Decode(b); err != nil {
		return xerrors.Errorf("failed to decode: %x", err)
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

func (c *Client) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(c.pingDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				c.log.Error("ping loop ctx error", zap.Error(err))
			}
			return
		case <-ticker.C:
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, c.pingTimeout)
				defer cancel()

				return c.Ping(ctx)
			}(); err != nil {
				c.log.Warn("ping error", zap.Error(err))
				if err := c.reconnect(); err != nil {
					c.log.Fatal("reconnect after ping error", zap.Error(err))
				}
			}
		}
	}
}
