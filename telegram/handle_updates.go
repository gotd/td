package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func (c *Client) processUpdates(updates tg.UpdatesClass) error {
	if c.updateHandler == nil {
		return nil
	}
	switch u := updates.(type) {
	case *tg.Updates:
		return c.updateHandler(c.ctx, u)
	case *tg.UpdateShort:
		// TODO(ernado): separate handler
		return c.updateHandler(c.ctx, &tg.Updates{
			Date: u.Date,
			Updates: []tg.UpdateClass{
				u.Update,
			},
		})
	// TODO(ernado): handle UpdatesTooLong
	// TODO(ernado): handle UpdateShortMessage
	// TODO(ernado): handle UpdateShortChatMessage
	// TODO(ernado): handle UpdatesCombined
	// TODO(ernado): handle UpdateShortSentMessage
	default:
		c.log.Warn("Ignoring update", zap.String("update_type", fmt.Sprintf("%T", u)))
	}
	return nil
}

func (c *Client) handleUpdates(b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return xerrors.Errorf("decode updates: %w", err)
	}
	return c.processUpdates(updates)
}
