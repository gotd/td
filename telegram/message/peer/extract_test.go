package peer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

type badPeer struct {
	tg.PeerClass
}

type mockResult struct {
	Entities
}

func (m mockResult) MapUsers() (r tg.UserClassArray) {
	for _, e := range m.Entities.Users() {
		r = append(r, e)
	}

	return r
}

func (m mockResult) MapChats() (r tg.ChatClassArray) {
	for _, e := range m.Entities.Chats() {
		r = append(r, e)
	}

	for _, e := range m.Entities.Channels() {
		r = append(r, e)
	}

	return r
}

func TestEntities(t *testing.T) {
	users := map[int64]*tg.User{
		10: {ID: 10, AccessHash: 10},
	}
	chats := map[int64]*tg.Chat{
		10: {ID: 10},
	}
	channels := map[int64]*tg.Channel{
		10: {ID: 10, AccessHash: 10},
	}
	ent := NewEntities(users, chats, channels)
	ctx := tg.Entities{
		Users:    users,
		Chats:    chats,
		Channels: channels,
	}
	result := mockResult{Entities: ent}

	tests := []struct {
		name   string
		filler func() Entities
	}{
		{"NewEntities", func() Entities {
			return ent
		}},
		{"EntitiesFromResult", func() Entities {
			return EntitiesFromResult(result)
		}},
		{"FillFromResult", func() Entities {
			e := NewEntities(
				map[int64]*tg.User{},
				map[int64]*tg.Chat{},
				map[int64]*tg.Channel{},
			)
			e.FillFromResult(result)
			return e
		}},
		{"EntitiesFromUpdate", func() Entities {
			return EntitiesFromUpdate(ctx)
		}},
		{"FillFromUpdate", func() Entities {
			e := NewEntities(
				map[int64]*tg.User{},
				map[int64]*tg.Chat{},
				map[int64]*tg.Channel{},
			)
			e.FillFromUpdate(ctx)
			return e
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)
			e := test.filler()

			_, err := e.ExtractPeer(badPeer{})
			a.Error(err)

			user, err := e.ExtractPeer(&tg.PeerUser{UserID: 10})
			peerUser := &tg.InputPeerUser{
				UserID:     10,
				AccessHash: 10,
			}
			a.Equal(peerUser, user)
			a.NoError(err)
			_, err = e.ExtractPeer(&tg.PeerUser{UserID: 11})
			a.Error(err)

			chat, err := e.ExtractPeer(&tg.PeerChat{ChatID: 10})
			peerChat := &tg.InputPeerChat{
				ChatID: 10,
			}
			a.Equal(peerChat, chat)
			a.NoError(err)
			_, err = e.ExtractPeer(&tg.PeerChat{ChatID: 11})
			a.Error(err)

			channel, err := e.ExtractPeer(&tg.PeerChannel{ChannelID: 10})
			peerChannel := &tg.InputPeerChannel{
				ChannelID:  10,
				AccessHash: 10,
			}
			a.Equal(peerChannel, channel)
			a.NoError(err)
			_, err = e.ExtractPeer(&tg.PeerChannel{ChannelID: 11})
			a.Error(err)
		})
	}
}
