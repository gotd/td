package updates

import (
	"context"

	"github.com/gotd/td/tg"
)

type update struct {
	Value interface{}
	State int
	Count int
	Ents  entities
	Ctx   context.Context
}

func (u update) start() int { return u.State - u.Count }

func (u update) end() int { return u.State }

// Entities contains update entities.
type entities struct {
	Users []tg.UserClass
	Chats []tg.ChatClass
}

// Merge merges entities.
func (e *entities) Merge(from entities) {
	for userIDFrom, userFrom := range from.Users {
		merge := true
		for userIDExist := range e.Users {
			if userIDExist == userIDFrom {
				merge = false
				break
			}
		}

		if merge {
			e.Users = append(e.Users, userFrom)
		}
	}

	for chatIDFrom, chatFrom := range from.Chats {
		merge := true
		for chatIDExist := range e.Chats {
			if chatIDExist == chatIDFrom {
				merge = false
				break
			}
		}

		if merge {
			e.Chats = append(e.Chats, chatFrom)
		}
	}
}
