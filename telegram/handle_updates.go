package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

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
	if v, ok := u.GetFwdFrom(); ok {
		msg.SetFwdFrom(v)
	}
	if v, ok := u.GetViaBotID(); ok {
		msg.SetViaBotID(v)
	}
	if v, ok := u.GetReplyTo(); ok {
		msg.SetReplyTo(v)
	}
	if v, ok := u.GetEntities(); ok {
		msg.SetEntities(v)
	}

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
	if v, ok := u.GetFwdFrom(); ok {
		msg.SetFwdFrom(v)
	}
	if v, ok := u.GetViaBotID(); ok {
		msg.SetViaBotID(v)
	}
	if v, ok := u.GetReplyTo(); ok {
		msg.SetReplyTo(v)
	}
	if v, ok := u.GetEntities(); ok {
		msg.SetEntities(v)
	}

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
	if v, ok := u.GetMedia(); ok {
		msg.SetMedia(v)
	}
	if v, ok := u.GetEntities(); ok {
		msg.SetEntities(v)
	}

	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      u.Pts,
			PtsCount: u.PtsCount,
		},
		Date: u.Date,
	}
}

func (c *Client) processUpdates(updates tg.UpdatesClass) error {
	if c.updateHandler == nil {
		return nil
	}
	switch u := updates.(type) {
	case *tg.Updates:
		return c.updateHandler.Handle(c.ctx, u)
	case *tg.UpdateShort:
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

func (c *Client) handleUpdates(b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return xerrors.Errorf("decode updates: %w", err)
	}
	return c.processUpdates(updates)
}
