package mtproto

import (
	"context"

	"github.com/gotd/td/bin"
)

func (c *Conn) write(ctx context.Context, id int64, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}

	c.mux.RLock()
	defer c.mux.RUnlock()
	if err := c.conn.Send(ctx, b); err != nil {
		return err
	}
	return nil
}

func (c *Conn) writeServiceMessage(ctx context.Context, message bin.Encoder) error {
	return c.write(ctx, c.newMessageID(), c.seqNo(), message)
}
