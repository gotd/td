package upconv

import "github.com/gotd/td/tg"

func convertOptional(msg *tg.Message, i tg.UpdatesClass) {
	if u, ok := i.(interface {
		GetFwdFrom() (tg.MessageFwdHeader, bool)
	}); ok {
		if v, ok := u.GetFwdFrom(); ok {
			msg.SetFwdFrom(v)
		}
	}
	if u, ok := i.(interface{ GetViaBotID() (int64, bool) }); ok {
		if v, ok := u.GetViaBotID(); ok {
			msg.SetViaBotID(v)
		}
	}
	if u, ok := i.(interface {
		GetReplyTo() (tg.MessageReplyHeader, bool)
	}); ok {
		if v, ok := u.GetReplyTo(); ok {
			msg.SetReplyTo(&v)
		}
	}
	if u, ok := i.(interface {
		GetReplyTo() (tg.MessageReplyHeaderClass, bool)
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

// ShortMessage converts UpdateShortMessage to UpdateShort.
func ShortMessage(u *tg.UpdateShortMessage) *tg.UpdateShort {
	msg := &tg.Message{
		ID:      u.ID,
		PeerID:  &tg.PeerUser{UserID: u.UserID},
		Message: u.Message,
		Date:    u.Date,
	}
	// Optional fields should set by SetXXX(), so GetXXX and Flags.Has()
	// can return the right values even we hav't call .Encode()
	msg.SetOut(u.Out)
	msg.SetMentioned(u.Mentioned)
	msg.SetMediaUnread(u.MediaUnread)
	msg.SetSilent(u.Silent)

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

// ShortChatMessage converts UpdateShortChatMessage to UpdateShort.
func ShortChatMessage(u *tg.UpdateShortChatMessage) *tg.UpdateShort {
	msg := &tg.Message{
		ID:      u.ID,
		PeerID:  &tg.PeerChat{ChatID: u.ChatID},
		Message: u.Message,
		Date:    u.Date,
	}

	msg.SetFromID(&tg.PeerUser{UserID: u.FromID})
	msg.SetOut(u.Out)
	msg.SetMentioned(u.Mentioned)
	msg.SetMediaUnread(u.MediaUnread)
	msg.SetSilent(u.Silent)

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

// ShortSentMessage converts UpdateShortSentMessage to UpdateShort.
func ShortSentMessage(u *tg.UpdateShortSentMessage) *tg.UpdateShort {
	msg := &tg.Message{
		ID:   u.ID,
		Date: u.Date,
	}
	msg.SetOut(u.Out)
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
