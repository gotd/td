package peers

import "github.com/gotd/td/tg"

type MsgAndEntities struct {
	Msg      string
	Entities []tg.MessageEntityClass
}
