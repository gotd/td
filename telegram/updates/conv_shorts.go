package updates

import "github.com/nnqq/td/tg"

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
	if u, ok := i.(interface {
		GetTTLPeriod() (int, bool)
	}); ok {
		if v, ok := u.GetTTLPeriod(); ok {
			msg.SetTTLPeriod(v)
		}
	}
}

func (s *state) convertShortMessage(u *tg.UpdateShortMessage) *tg.UpdateShort {
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

	msg.SetFromID(&tg.PeerUser{UserID: s.selfID})
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

func (s *state) convertShortChatMessage(u *tg.UpdateShortChatMessage) *tg.UpdateShort {
	msg := &tg.Message{
		ID:      u.ID,
		PeerID:  &tg.PeerChat{ChatID: u.ChatID},
		Message: u.Message,
		Date:    u.Date,
	}

	msg.SetOut(u.Out)
	msg.SetMentioned(u.Mentioned)
	msg.SetMediaUnread(u.MediaUnread)
	msg.SetSilent(u.Silent)
	msg.SetFromScheduled(msg.FromScheduled)
	msg.SetFromID(&tg.PeerUser{UserID: u.FromID})
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

func (s *state) convertShortSentMessage(u *tg.UpdateShortSentMessage) *tg.UpdateShort {
	// This update should be converted by the one who called the method
	// that returned this update, because we do not have any context about
	// it (message text, sender/recipient, etc.)
	//
	// In theory, this update should come only as a response to an RPC call,
	// and we get it here because of the update hook.
	// We use it to make sure there are no pts gaps.
	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message: &tg.MessageEmpty{
				ID: u.ID,
			},
			Pts:      u.Pts,
			PtsCount: u.PtsCount,
		},
		Date: u.Date,
	}
}
