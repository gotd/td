package e2e

import (
	"sync"

	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

var _ updates.Handler = (*Handler)(nil)

// Handler handles updates.
type Handler struct {
	messages *messageDatabase
	ents     *updates.Entities
	mux      sync.Mutex
}

// NewHandler creates new update handler.
func NewHandler() *Handler {
	return &Handler{
		messages: &messageDatabase{
			channels: make(map[int][]tg.MessageClass),
		},
		ents: updates.NewEntities(),
	}
}

// ChannelTooLong handler.
func (h *Handler) ChannelTooLong(channelID int) {
	panic("not implemented")
}

// HandleDiff handler.
func (h *Handler) HandleDiff(diff updates.DiffUpdate) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	for _, msg := range diff.NewMessages {
		if channel, ok := msg.(*tg.Message).PeerID.(*tg.PeerChannel); ok {
			msgs := h.messages.channels[channel.ChannelID]
			msgs = append(msgs, msg)
			h.messages.channels[channel.ChannelID] = msgs
			continue
		}

		h.messages.common = append(h.messages.common, msg)
	}

	h.messages.secret = append(h.messages.secret, diff.NewEncryptedMessages...)
	for _, u := range diff.Users {
		switch u := u.(type) {
		case *tg.User:
			h.ents.Users[u.ID] = u
		default:
			panic("bad user type")
		}
	}

	for _, c := range diff.Chats {
		switch c := c.(type) {
		case *tg.Chat:
			h.ents.Chats[c.ID] = c
		case *tg.Channel:
			h.ents.Channels[c.ID] = c
		default:
			panic("bad chat type")
		}
	}
	return nil
}

// HandleUpdates handler.
func (h *Handler) HandleUpdates(ents *updates.Entities, upds []tg.UpdateClass) error {
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
