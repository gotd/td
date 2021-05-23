package telegram

import (
	"context"

	"github.com/gotd/td/tg"
)

// SendMessage sends message to peer.
//
// Deprecated: use helpers like message.NewSender.
func (c *Client) SendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
	if req.RandomID == 0 {
		id, err := c.RandInt64()
		if err != nil {
			return err
		}
		req.RandomID = id
	}
	updates, err := c.tg.MessagesSendMessage(ctx, req)
	if err != nil {
		return err
	}
	return c.processUpdates(updates)
}
