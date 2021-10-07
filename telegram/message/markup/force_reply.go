package markup

import "github.com/nnqq/td/tg"

// ForceReply creates markup to force the user to send a reply.
func ForceReply(singleUse, selective bool) tg.ReplyMarkupClass {
	return &tg.ReplyKeyboardForceReply{
		SingleUse: singleUse,
		Selective: selective,
	}
}
