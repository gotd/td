package telegram

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/internal/proto"
)

func (c *Client) write(id crypto.MessageID, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return err
	}
	return nil
}
