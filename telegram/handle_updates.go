package telegram

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func convertOptional(msg *tg.Message, i tg.UpdatesClass) {
	if u, ok := i.(interface {
		GetFwdFrom() (tg.MessageFwdHeader, bool)
	}); ok {
		if v, ok := u.GetFwdFrom(); ok {
			msg.SetFwdFrom(v)
		}
	}
	if u, ok := i.(interface{ GetViaBotID() (int, bool) }); ok {
		if v, ok := u.GetViaBotID(); ok {
			msg.SetViaBotID(v)
		}
	}
	if u, ok := i.(interface {
		GetReplyTo() (tg.MessageReplyHeader, bool)
	}); ok {
		if v, ok := u.GetReplyTo(); ok {
			msg.SetReplyTo(v)
		}
	}
	if u, ok := i.(interface {
		GetEntities() ([]tg.MessageEntityClass, bool)
	}); ok {
		if v, ok := u.GetEntities(); ok {
			msg.SetEntities(v)
		}
	}
	if u, ok := i.(interface {
		GetMedia() (tg.MessageMediaClass, bool)
	}); ok {
		if v, ok := u.GetMedia(); ok {
			msg.SetMedia(v)
		}
	}
}

func convertUpdateShortMessage(u *tg.UpdateShortMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out:         u.Out,
		Mentioned:   u.Mentioned,
		MediaUnread: u.MediaUnread,
		Silent:      u.Silent,
		ID:          u.ID,
		FromID:      &tg.PeerUser{UserID: u.UserID},
		PeerID:      &tg.PeerUser{UserID: u.UserID},
		Message:     u.Message,
	}
	convertOptional(msg, u)

	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      u.Pts,
			PtsCount: u.PtsCount,
		},
		Date: u.Date,
	}
}

func convertUpdateShortChatMessage(u *tg.UpdateShortChatMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out:         u.Out,
		Mentioned:   u.Mentioned,
		MediaUnread: u.MediaUnread,
		Silent:      u.Silent,
		ID:          u.ID,
		FromID:      &tg.PeerUser{UserID: u.FromID},
		PeerID:      &tg.PeerChat{ChatID: u.ChatID},
		Message:     u.Message,
	}
	convertOptional(msg, u)

	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      u.Pts,
			PtsCount: u.PtsCount,
		},
		Date: u.Date,
	}
}

func convertUpdateShortSentMessage(u *tg.UpdateShortSentMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out: u.Out,
		ID:  u.ID,
	}
	convertOptional(msg, u)

	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      u.Pts,
			PtsCount: u.PtsCount,
		},
		Date: u.Date,
	}
}

func (c *Client) updateInterceptor(updates ...tg.UpdateClass) {
	for _, update := range updates {
		_, ok := update.(*tg.UpdateConfig)
		if ok {
			cfg, err := c.tg.HelpGetConfig(c.ctx)
			if err != nil {
				c.log.Warn("Fetch config", zap.Error(err))
				continue
			}

			if err := c.onPrimaryConfig(*cfg); err != nil {
				c.log.Warn("Save config", zap.Error(err))
				continue
			}
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
		return c.processUpdates(convertUpdateShortMessage(u))
	case *tg.UpdateShortChatMessage:
		return c.processUpdates(convertUpdateShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return c.processUpdates(convertUpdateShortSentMessage(u))
	// TODO(ernado): handle UpdatesTooLong
	// TODO(ernado): handle UpdatesCombined
	default:
		c.log.Warn("Ignoring update", zap.String("update_type", fmt.Sprintf("%T", u)))
	}
	return nil
}
