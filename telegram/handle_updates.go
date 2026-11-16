package telegram

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/telegram/internal/upconv"
	"github.com/gotd/td/tg"
)

func (c *Client) updateInterceptor(updates ...tg.UpdateClass) {
	for _, update := range updates {
		switch u := update.(type) {
		case *tg.UpdateConfig, *tg.UpdateDCOptions:
			c.fetchConfig(c.ctx)
		case *tg.UpdateServiceNotification:
			c.handleServiceNotification(u)
		}
	}
}

func (c *Client) processUpdates(updates tg.UpdatesClass) error {
	switch u := updates.(type) {
	case *tg.Updates:
		c.updateInterceptor(u.Updates...)
		return c.updateHandler.Handle(c.ctx, u)
	case *tg.UpdatesCombined:
		c.updateInterceptor(u.Updates...)
		return c.updateHandler.Handle(c.ctx, u)
	case *tg.UpdateShort:
		c.updateInterceptor(u.Update)
		return c.updateHandler.Handle(c.ctx, u)
	case *tg.UpdateShortMessage:
		return c.updateHandler.Handle(c.ctx, upconv.ShortMessage(u))
	case *tg.UpdateShortChatMessage:
		return c.updateHandler.Handle(c.ctx, upconv.ShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return c.updateHandler.Handle(c.ctx, upconv.ShortSentMessage(u))
	case *tg.UpdatesTooLong:
		return c.updateHandler.Handle(c.ctx, u)
	default:
		c.log.Warn(c.ctx, "Ignoring update", log.String("update_type", fmt.Sprintf("%T", u)))
	}
	return nil
}

func (c *Client) handleUpdates(b *bin.Buffer) error {
	return c.handleUpdatesInfo(mtproto.MessageInfo{}, b)
}

// handleUpdatesInfo handles updates pushed by the server and records their
// timings, see updateTiming.
func (c *Client) handleUpdatesInfo(info mtproto.MessageInfo, b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return errors.Wrap(err, "decode updates")
	}
	if info.Received.IsZero() {
		// Not from a connection (tests).
		return c.processUpdates(updates)
	}

	start := c.now()
	err = c.processUpdates(updates)
	c.meters().updateTiming(c.ctx, c.log, updateTiming{
		created:  proto.MessageID(info.ID).Time(),
		received: info.Received,
		offset:   info.ServerTimeOffset,
		start:    start,
		done:     c.now(),
		failed:   err != nil,
	}, updates)
	return err
}
