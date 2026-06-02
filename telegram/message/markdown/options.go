package markdown

import (
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

// Options is options of Markdown parser.
type Options struct {
	// UserResolver is used to resolve user by ID during formatting. May be nil.
	//
	// If userResolver is nil, formatter will create tg.InputUser using only ID.
	// Notice that it's okay for bots, but not for users.
	UserResolver entity.UserResolver
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
