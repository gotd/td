package e2e

import (
	"github.com/nnqq/td/tg"
)

// EventBuilder struct.
type EventBuilder struct {
	updates []tg.UpdateClass
	ents    *Entities
	s       *server
	date    int
}

// SendMessage send a new message.
func (e *EventBuilder) SendMessage(from *tg.PeerUser, peer tg.PeerClass, text string) {
	msg := &tg.Message{
		Message: text,
		PeerID:  peer,
		FromID:  from,
		Date:    e.date,
	}

	fromUser, ok := e.s.peers.users[from.UserID]
	if !ok {
		panic("bad fromID")
	}
	e.ents.Users[from.UserID] = fromUser

	switch peer := peer.(type) {
	case *tg.PeerUser:
		user, ok := e.s.peers.users[peer.UserID]
		if !ok {
			panic("peer not found")
		}

		e.ents.Users[user.ID] = user
		e.s.messages.common = append(e.s.messages.common, msg)
		e.updates = append(e.updates, &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      len(e.s.messages.common),
			PtsCount: 1,
		})
	case *tg.PeerChat:
		chat, ok := e.s.peers.chats[peer.ChatID]
		if !ok {
			panic("peer not found")
		}

		e.ents.Chats[chat.ID] = chat
		e.s.messages.common = append(e.s.messages.common, msg)
		e.updates = append(e.updates, &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      len(e.s.messages.common),
			PtsCount: 1,
		})
	case *tg.PeerChannel:
		channel, ok := e.s.peers.channels[peer.ChannelID]
		if !ok {
			panic("peer not found")
		}

		e.ents.Channels[channel.ID] = channel
		msgs := append(e.s.messages.channels[peer.ChannelID], msg)
		e.s.messages.channels[peer.ChannelID] = msgs
		e.updates = append(e.updates, &tg.UpdateNewChannelMessage{
			Message:  msg,
			Pts:      len(msgs),
			PtsCount: 1,
		})
	default:
		panic("unexpected peer type")
	}
}

// CreateEvent creates new event.
func (s *server) CreateEvent(f func(ev *EventBuilder)) *tg.Updates {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.date++
	ev := &EventBuilder{
		ents: NewEntities(),
		s:    s,
		date: s.date,
	}
	f(ev)

	return &tg.Updates{
		Updates: ev.updates,
		Users:   ev.ents.AsUsers(),
		Chats:   ev.ents.AsChats(),
		Date:    s.date,
		Seq:     0,
	}
}
