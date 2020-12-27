package mtproto

import (
	"context"

	"github.com/gotd/td/bin"
)

func (c *Client) write(ctx context.Context, id int64, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}

	c.connMux.RLock()
	defer c.connMux.RUnlock()
	if err := c.conn.Send(ctx, b); err != nil {
		return err
	}
	return nil
}

func (c *Client) writeServiceMessage(ctx context.Context, message bin.Encoder) error {
	return c.write(ctx, c.newMessageID(), c.seqNo(), message)
}
