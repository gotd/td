package telegram

import (
	"context"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/tg"
)

type updatesBox struct {
	Updates tg.UpdatesClass
}

func (u *updatesBox) Decode(b *bin.Buffer) error {
	v, err := tg.DecodeUpdates(b)
	if err != nil {
		return err
	}
	u.Updates = v
	return nil
}

// SendMessage sends message to peer.
func (c *Client) SendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
	var res updatesBox
	if err := c.do(ctx, req, &res); err != nil {
		return err
	}
	return c.processUpdates(res.Updates)
}
