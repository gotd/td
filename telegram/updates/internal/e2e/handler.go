package e2e

import (
	"context"
	"sync"

	"github.com/nnqq/td/tg"
)

// Handler handles updates.
type handler struct {
	messages *messageDatabase
	ents     *Entities
	mux      sync.Mutex
}

func newHandler() *handler {
	return &handler{
		messages: &messageDatabase{
			channels: make(map[int64][]tg.MessageClass),
		},
		ents: NewEntities(),
	}
}

func (h *handler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	switch u := u.(type) {
	case *tg.Updates:
		return h.handleUpdates(NewEntities().FromUpdates(u), u.Updates)
	case *tg.UpdatesCombined:
		return h.handleUpdates(NewEntities().FromUpdates(u), u.Updates)
	default:
		panic(u)
	}
}

// HandleUpdates handler.
func (h *handler) handleUpdates(ents *Entities, upds []tg.UpdateClass) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.ents.Merge(ents)
	for _, u := range upds {
		switch u := u.(type) {
		case *tg.UpdateNewMessage:
			h.messages.common = append(h.messages.common, u.Message)
		case *tg.UpdateNewEncryptedMessage:
			h.messages.secret = append(h.messages.secret, u.Message)
		case *tg.UpdateNewChannelMessage:
			channelID := u.Message.(*tg.Message).PeerID.(*tg.PeerChannel).ChannelID
			msgs := h.messages.channels[channelID]
			msgs = append(msgs, u.Message)
			h.messages.channels[channelID] = msgs
		default:
			panic("unexpected update type")
		}
	}

	return nil
}
