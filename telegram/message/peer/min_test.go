package peer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestEntitiesMinPeers(t *testing.T) {
	ent := NewEntities(map[int64]*tg.User{
		1: {ID: 1, AccessHash: 11},
		2: {ID: 2, Min: true, AccessHash: 22},
	}, map[int64]*tg.Chat{}, map[int64]*tg.Channel{
		10: {ID: 10, AccessHash: 100},
		20: {ID: 20, Min: true, AccessHash: 200},
	})

	p, err := ent.ExtractPeer(&tg.PeerUser{UserID: 1})
	require.NoError(t, err)
	require.Equal(t, &tg.InputPeerUser{UserID: 1, AccessHash: 11}, p)

	_, err = ent.ExtractPeer(&tg.PeerUser{UserID: 2})
	require.ErrorIs(t, err, ErrMinPeer)
	_, err = ent.ExtractPeer(&tg.PeerChannel{ChannelID: 20})
	require.ErrorIs(t, err, ErrMinPeer)

	chat := &tg.InputPeerChannel{ChannelID: 10, AccessHash: 100}
	p, err = ent.ExtractPeerFromMessage(&tg.PeerUser{UserID: 2}, chat, 5)
	require.NoError(t, err)
	require.Equal(t, &tg.InputPeerUserFromMessage{Peer: chat, MsgID: 5, UserID: 2}, p)
	p, err = ent.ExtractPeerFromMessage(&tg.PeerChannel{ChannelID: 20}, chat, 6)
	require.NoError(t, err)
	require.Equal(t, &tg.InputPeerChannelFromMessage{Peer: chat, MsgID: 6, ChannelID: 20}, p)
	p, err = ent.ExtractPeerFromMessage(&tg.PeerUser{UserID: 1}, chat, 5)
	require.NoError(t, err)
	require.Equal(t, &tg.InputPeerUser{UserID: 1, AccessHash: 11}, p)
	_, err = ent.ExtractPeerFromMessage(&tg.PeerUser{UserID: 3}, chat, 5)
	require.Error(t, err)

	// Min entities do not replace full ones.
	ent.Fill(map[int64]*tg.User{
		1: {ID: 1, Min: true},
		2: {ID: 2, AccessHash: 23},
	}, nil, map[int64]*tg.Channel{
		10: {ID: 10, Min: true},
	})
	require.False(t, ent.users[1].Min)
	require.Equal(t, int64(11), ent.users[1].AccessHash)
	require.Equal(t, int64(23), ent.users[2].AccessHash, "full replaces min")
	require.False(t, ent.channels[10].Min)

	ent.FillFromResult(&tg.MessagesMessages{Users: []tg.UserClass{&tg.User{ID: 1, Min: true}}})
	require.False(t, ent.users[1].Min)
}
