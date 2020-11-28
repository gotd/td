package telegram

import "github.com/ernado/td/crypto"

func (c *Client) newMessageID() crypto.MessageID {
	return crypto.NewMessageID(c.clock(), crypto.MessageFromClient)
}
