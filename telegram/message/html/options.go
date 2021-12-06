package html

import (
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

// Options is options of HTML.
type Options struct {
	// UserResolver is used to resolve user by ID during formatting. May be nil.
	//
	// If userResolver is nil, formatter will create tg.InputUser using only ID.
	// Notice that it's okay for bots, but not for users.
	UserResolver entity.UserResolver
	// DisableTelegramEscape disable Telegram BotAPI escaping and uses default
	// golang.org/x/net/html escape.
	DisableTelegramEscape bool
}

func (o *Options) setDefaults() {
	if o.UserResolver == nil {
		o.UserResolver = func(id int64) (tg.InputUserClass, error) {
			return &tg.InputUser{
				UserID: id,
			}, nil
		}
	}
}
