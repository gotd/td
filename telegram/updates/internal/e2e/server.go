// Package e2e contains end-to-end updates processing test.
package e2e

import (
	"context"
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// Server for testing gaps.
type server struct {
	date     int
	peers    *peerDatabase
	messages *messageDatabase

	mux sync.Mutex
}

// NewServer creates new test server.
func newServer() *server {
	return &server{
		date: 1,
		peers: &peerDatabase{
			users:    make(map[int64]*tg.User),
			chats:    make(map[int64]*tg.Chat),
			channels: make(map[int64]*tg.Channel),
		},
		messages: &messageDatabase{
			channels: make(map[int64][]tg.MessageClass),
		},
	}
}

// UpdatesGetState returns current remote state.
func (s *server) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return &tg.UpdatesState{
		Pts:  len(s.messages.common),
		Qts:  len(s.messages.secret),
		Date: s.date,
		Seq:  0,
	}, nil
}

// UpdatesGetDifference returns difference between local and remote states.
func (s *server) UpdatesGetDifference(ctx context.Context, request *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	ents := NewEntities()
	var common []tg.MessageClass
	for i := request.Pts + 1; i <= len(s.messages.common); i++ {
		common = append(common, s.messages.common[i-1])
		s.fillMessageEnts(s.messages.common[i-1], ents)
	}

	var secret []tg.EncryptedMessageClass
	for i := request.Qts + 1; i <= len(s.messages.secret); i++ {
		secret = append(secret, s.messages.secret[i-1])
	}

	var others []tg.UpdateClass
	for _, msgs := range s.messages.channels {
		for i, msg := range msgs {
			if msg.(*tg.Message).Date > request.Date {
				others = append(others, &tg.UpdateNewChannelMessage{
					Message:  msg,
					Pts:      i + 1,
					PtsCount: 1,
				})
				s.fillMessageEnts(msg, ents)
			}
		}
	}

	if len(common) == 0 && len(secret) == 0 && len(others) == 0 {
		return &tg.UpdatesDifferenceEmpty{
			Date: s.date,
			Seq:  0,
		}, nil
	}

	return &tg.UpdatesDifference{
		NewMessages:          common,
		NewEncryptedMessages: secret,
		OtherUpdates:         others,
		Users:                ents.AsUsers(),
		Chats:                ents.AsChats(),
		State: tg.UpdatesState{
			Pts:  len(s.messages.common),
			Qts:  len(s.messages.secret),
			Date: s.date,
			Seq:  0,
		},
	}, nil
}

// UpdatesGetChannelDifference returns difference between local and remote channel states.
func (s *server) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	channel, ok := request.Channel.(*tg.InputChannel)
	if !ok {
		return nil, xerrors.Errorf("bad InputChannelClass type: %T", request.Channel)
	}

	if peer, ok := s.peers.channels[channel.ChannelID]; true {
		if !ok {
			return nil, xerrors.Errorf("channel %d not found", channel.ChannelID)
		}

		if peer.AccessHash != channel.AccessHash {
			return nil, xerrors.New("invalid access hash")
		}
	}

	var (
		channelMsgs = s.messages.channels[channel.ChannelID]
		ents        = NewEntities()
		prepared    []tg.MessageClass
	)

	for i := request.Pts + 1; i <= len(channelMsgs); i++ {
		prepared = append(prepared, channelMsgs[i-1])
		s.fillMessageEnts(channelMsgs[i-1], ents)
	}

	if len(prepared) == 0 {
		return &tg.UpdatesChannelDifferenceEmpty{
			Pts:   len(channelMsgs),
			Final: true,
		}, nil
	}

	return &tg.UpdatesChannelDifference{
		NewMessages: prepared,
		Users:       ents.AsUsers(),
		Chats:       ents.AsChats(),
		Pts:         len(channelMsgs),
		Final:       true,
	}, nil
}

func (s *server) fillMessageEnts(msg tg.MessageClass, ents *Entities) {
	switch peer := msg.(*tg.Message).PeerID.(type) {
	case *tg.PeerUser:
		user, ok := s.peers.users[peer.UserID]
		if !ok {
			panic("bad user")
		}

		ents.Users[user.ID] = user
	case *tg.PeerChat:
		chat, ok := s.peers.chats[peer.ChatID]
		if !ok {
			panic("bad chat")
		}

		ents.Chats[chat.ID] = chat
	case *tg.PeerChannel:
		channel, ok := s.peers.channels[peer.ChannelID]
		if !ok {
			panic("bad channel")
		}

		ents.Channels[channel.ID] = channel
	default:
		panic("unexpected peer type")
	}

	peerUser, ok := msg.(*tg.Message).FromID.(*tg.PeerUser)
	if !ok {
		panic("bad fromID")
	}

	user, ok := s.peers.users[peerUser.UserID]
	if !ok {
		panic("bad user")
	}

	ents.Users[user.ID] = user
}
