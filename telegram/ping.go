package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/crypto"
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

	if err := c.write(c.newMessageID(), pingMessage{id: pingID}); err != nil {
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
