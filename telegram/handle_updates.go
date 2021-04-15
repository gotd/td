package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/helpers"
	"github.com/gotd/td/tg"
)

func (c *Client) updateInterceptor(updates ...tg.UpdateClass) {
	for _, update := range updates {
		switch update.(type) {
		case *tg.UpdateConfig, *tg.UpdateDCOptions:
			c.fetchConfig(c.ctx)
		}
	}
}

func (c *Client) processUpdates(updates tg.UpdatesClass) error {
	switch u := updates.(type) {
	case *tg.Updates:
		c.updateInterceptor(u.Updates...)
		if c.updateHandler == nil {
			return nil
		}
		return c.updateHandler.Handle(c.ctx, u)
	case *tg.UpdateShort:
		c.updateInterceptor(u.Update)
		if c.updateHandler == nil {
			return nil
		}
		return c.updateHandler.HandleShort(c.ctx, u)
	case *tg.UpdateShortMessage:
		return c.processUpdates(helpers.ConvertUpdateShortMessage(u))
	case *tg.UpdateShortChatMessage:
		return c.processUpdates(helpers.ConvertUpdateShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return c.processUpdates(helpers.ConvertUpdateShortSentMessage(u))
	// TODO(ernado): handle UpdatesTooLong
	// TODO(ernado): handle UpdatesCombined
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
