package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
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
	updateCfg := false

	for _, update := range updates {
		switch update.(type) {
		case *tg.UpdateConfig, *tg.UpdateDCOptions:
			updateCfg = true
		}
	}

	if updateCfg {
		c.fetchConfig(c.ctx)
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
	case *tg.UpdatesCombined:
		c.updateInterceptor(u.Updates...)
		if c.updateHandler == nil {
			return nil
		}
		return c.updateHandler.Handle(c.ctx, &tg.Updates{
			Updates: u.Updates,
			Users:   u.Users,
			Chats:   u.Chats,
			Date:    u.Date,
			Seq:     u.Seq,
		})
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
	default:
		// We ignoring UpdatesTooLong because we should not receive it here.
		// It used only in explicit update requests.
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
