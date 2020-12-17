package telegram

import (
	"github.com/gotd/td/bin"
)

func (c *Client) write(id int64, seq int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(id, seq, message, b); err != nil {
		return err
	}

	if isPriorityMessage(message) {
		c.pchan <- b
		return nil
	}

	c.wchan <- b
	return nil
}
