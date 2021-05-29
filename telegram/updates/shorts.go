package updates

import "github.com/gotd/td/tg"

func (e *Engine) convertShortMessage(short *tg.UpdateShortMessage) *tg.UpdateShort {
	fromID := short.UserID
	if short.Out {
		fromID = e.selfID
	}

	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message: &tg.Message{
				Out:         short.Out,
				Mentioned:   short.Mentioned,
				MediaUnread: short.MediaUnread,
				Silent:      short.Silent,
				ID:          short.ID,
				FromID: &tg.PeerUser{
					UserID: fromID,
				},
				PeerID: &tg.PeerChat{
					ChatID: short.UserID,
				},
				Message:   short.Message,
				Date:      short.Date,
				FwdFrom:   short.FwdFrom,
				ViaBotID:  short.ViaBotID,
				ReplyTo:   short.ReplyTo,
				Entities:  short.Entities,
				TTLPeriod: short.TTLPeriod,
			},
			Pts:      short.Pts,
			PtsCount: short.PtsCount,
		},
		Date: short.Date,
	}
}

func (e *Engine) convertShortChatMessage(short *tg.UpdateShortChatMessage) *tg.UpdateShort {
	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message: &tg.Message{
				Out:         short.Out,
				Mentioned:   short.Mentioned,
				MediaUnread: short.MediaUnread,
				Silent:      short.Silent,
				ID:          short.ID,
				FromID: &tg.PeerUser{
					UserID: short.FromID,
				},
				PeerID: &tg.PeerChat{
					ChatID: short.ChatID,
				},
				Message:   short.Message,
				Date:      short.Date,
				FwdFrom:   short.FwdFrom,
				ViaBotID:  short.ViaBotID,
				ReplyTo:   short.ReplyTo,
				Entities:  short.Entities,
				TTLPeriod: short.TTLPeriod,
			},
			Pts:      short.Pts,
			PtsCount: short.PtsCount,
		},
		Date: short.Date,
	}
}

func (e *Engine) convertShortSentMessage(short *tg.UpdateShortSentMessage) *tg.UpdateShort {
	return &tg.UpdateShort{
		Update: &tg.UpdateNewMessage{
			Message: &tg.MessageEmpty{
				ID: short.ID,
			},
			Pts:      short.Pts,
			PtsCount: short.PtsCount,
		},
		Date: short.Date,
	}
}
