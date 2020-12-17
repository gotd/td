package telegram

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
)

func (c *Client) write(id int64, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return err
	}
	return nil
}

func (c *Client) writeServiceMessage(message bin.Encoder) error {
	return c.write(c.newMessageID(), c.seqNo(), message)
}
