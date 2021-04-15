package helpers

import "github.com/gotd/td/tg"

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

// ConvertUpdateShortMessage converts UpdateShortMessage to UpdateShort.
func ConvertUpdateShortMessage(u *tg.UpdateShortMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out:         u.Out,
		Mentioned:   u.Mentioned,
		MediaUnread: u.MediaUnread,
		Silent:      u.Silent,
		ID:          u.ID,
		PeerID:      &tg.PeerUser{UserID: u.UserID},
		Message:     u.Message,
		Date:        u.Date,
	}
	if !u.Out {
		msg.SetFromID(&tg.PeerUser{UserID: u.UserID})
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

// ConvertUpdateShortChatMessage converts UpdateShortChatMessage to UpdateShort.
func ConvertUpdateShortChatMessage(u *tg.UpdateShortChatMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out:         u.Out,
		Mentioned:   u.Mentioned,
		MediaUnread: u.MediaUnread,
		Silent:      u.Silent,
		ID:          u.ID,
		FromID:      &tg.PeerUser{UserID: u.FromID},
		PeerID:      &tg.PeerChat{ChatID: u.ChatID},
		Message:     u.Message,
		Date:        u.Date,
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

// ConvertUpdateShortSentMessage converts UpdateShortSentMessage to UpdateShort.
func ConvertUpdateShortSentMessage(u *tg.UpdateShortSentMessage) *tg.UpdateShort {
	msg := &tg.Message{
		Out:  u.Out,
		ID:   u.ID,
		Date: u.Date,
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
