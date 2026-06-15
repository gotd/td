package updates

import (
	"github.com/gotd/td/tg"
)

type update struct {
	Value    any
	State    int
	Count    int
	Entities entities
}

// affectedPts is a synthetic, non-dispatchable pts update. It carries only the
// pts increment from a messages.affectedMessages / messages.affectedHistory RPC
// result so the local pts stays in sync with the server after a self-initiated
// read or delete, without fabricating a user-visible update.
//
// It is fed to a sequenceBox as update.Value; applyPts skips it instead of
// dispatching it to the handler. See Manager.HandleAffected.
type affectedPts struct{}

func (u update) start() int { return u.State - u.Count }

func (u update) end() int { return u.State }

// Entities contains update entities.
type entities struct {
	Users []tg.UserClass
	Chats []tg.ChatClass
}

// Merge merges entities.
func (e *entities) Merge(from entities) {
	for _, candidate := range from.Users {
		merge := true
		for _, exist := range e.Users {
			if exist.GetID() == candidate.GetID() {
				merge = false
				break
			}
		}

		if merge {
			e.Users = append(e.Users, candidate)
		}
	}

	for _, candidate := range from.Chats {
		merge := true
		for _, exist := range e.Chats {
			if exist.GetID() == candidate.GetID() {
				merge = false
				break
			}
		}

		if merge {
			e.Chats = append(e.Chats, candidate)
		}
	}
}
