package telegram

import (
	"context"

	"github.com/gotd/td/bin"
)

func (c *Client) write(ctx context.Context, id int64, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}
	if err := c.conn.Send(ctx, b); err != nil {
		return err
	}
	return nil
}
