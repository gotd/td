package telegram

import (
	"context"

	"github.com/gotd/td/tg"
)

// SendMessage sends message to peer.
func (c *Client) SendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
	var res tg.UpdatesBox
	if err := c.rpcContent(ctx, req, &res); err != nil {
		return err
	}
	return c.processUpdates(res.Updates)
}
