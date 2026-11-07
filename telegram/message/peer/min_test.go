package peer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// flagged returns v with flags set, as a decoded object has them.
func flagged[T interface{ SetFlags() }](v T) T {
	v.SetFlags()
	return v
}

// roundTrip encodes v and decodes it into out.
func roundTrip(t *testing.T, v, out bin.Object) {
	t.Helper()
	var b bin.Buffer
	require.NoError(t, v.Encode(&b))
	require.NoError(t, out.Decode(&b))
}

func TestMergeMinFlags(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		full := flagged(&tg.User{
			ID: 1, AccessHash: 11, FirstName: "Full",
			EmojiStatus: &tg.EmojiStatus{DocumentID: 7},
		})
		minUser := flagged(&tg.User{ID: 1, Min: true, Premium: true})

		u := MergeMinUser(full, minUser)
		require.True(t, u.GetPremium(), "getter must see merged flag")
		_, ok := u.GetEmojiStatus()
		require.False(t, ok, "emoji status removed by min must not be reported")

		var decoded tg.User
		roundTrip(t, u, &decoded)
		require.True(t, decoded.Premium)
		require.Nil(t, decoded.EmojiStatus)
		require.Equal(t, "Full", decoded.FirstName)
		require.Equal(t, int64(11), decoded.AccessHash)
	})
	t.Run("Channel", func(t *testing.T) {
		full := flagged(&tg.Channel{
			ID: 10, AccessHash: 100, Title: "Old",
			Photo:       &tg.ChatPhotoEmpty{},
			EmojiStatus: &tg.EmojiStatus{DocumentID: 7},
		})
		minChannel := flagged(&tg.Channel{
			ID: 10, Min: true, Title: "New", Verified: true,
			Photo:     &tg.ChatPhotoEmpty{},
			Usernames: []tg.Username{{Username: "new", Active: true}},
		})

		c := MergeMinChannel(full, minChannel)
		require.True(t, c.GetVerified())
		_, ok := c.GetEmojiStatus()
		require.False(t, ok)
		usernames, ok := c.GetUsernames()
		require.True(t, ok)
		require.Len(t, usernames, 1)

		var decoded tg.Channel
		roundTrip(t, c, &decoded)
		require.Equal(t, "New", decoded.Title)
		require.True(t, decoded.Verified)
		require.Nil(t, decoded.EmojiStatus)
		require.Equal(t, int64(100), decoded.AccessHash)
	})
}

func TestMergeMinChannelTDLib(t *testing.T) {
	color := &tg.PeerColor{}
	color.SetColor(3)
	fullPhoto := &tg.ChatPhoto{PhotoID: 1}
	minPhoto := &tg.ChatPhoto{PhotoID: 2}
	full := &tg.Channel{ID: 10, AccessHash: 100, Photo: fullPhoto, Color: color}
	minChannel := &tg.Channel{ID: 10, Min: true, Photo: minPhoto}

	c := MergeMinChannel(full, minChannel)
	require.Equal(t, color, c.Color, "colors are not taken from min")
	require.Equal(t, minPhoto, c.Photo)

	banned := *full
	banned.BannedRights = tg.ChatBannedRights{ViewMessages: true}
	c = MergeMinChannel(&banned, minChannel)
	require.Equal(t, fullPhoto, c.Photo, "photo is kept when banned")
}

func TestMergeMinUserDeletedSelf(t *testing.T) {
	full := &tg.User{ID: 1, Self: true, AccessHash: 11, FirstName: "Me"}
	u := MergeMinUser(full, &tg.User{ID: 1, Min: true, Deleted: true, Premium: true})
	require.False(t, u.Deleted, "deleted min self is ignored")
	require.False(t, u.Premium)
	require.Equal(t, "Me", u.FirstName)

	other := &tg.User{ID: 2, AccessHash: 22}
	require.True(t, MergeMinUser(other, &tg.User{ID: 2, Min: true, Deleted: true}).Deleted)
}

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

func TestEntitiesFillMergesMin(t *testing.T) {
	fullPhoto := &tg.UserProfilePhoto{PhotoID: 1}
	minPhoto := &tg.UserProfilePhoto{PhotoID: 2}
	ent := NewEntities(map[int64]*tg.User{
		1: {
			ID: 1, AccessHash: 11, FirstName: "Full", Phone: "123",
			Photo: fullPhoto, Status: &tg.UserStatusRecently{},
		},
		2: {ID: 2, AccessHash: 22, ApplyMinPhoto: true, Photo: fullPhoto},
	}, map[int64]*tg.Chat{}, map[int64]*tg.Channel{
		10: {ID: 10, AccessHash: 100, Title: "Old", Creator: true, ParticipantsCount: 5},
	})

	ent.Fill(map[int64]*tg.User{
		1: {
			ID: 1, Min: true, AccessHash: 99, FirstName: "Min", Premium: true,
			Photo: minPhoto, Status: &tg.UserStatusOffline{},
			EmojiStatus: &tg.EmojiStatus{DocumentID: 7},
		},
		2: {ID: 2, Min: true, AccessHash: 98, Photo: minPhoto},
	}, nil, map[int64]*tg.Channel{
		10: {ID: 10, Min: true, AccessHash: 999, Title: "New", Verified: true},
	})

	u := ent.users[1]
	require.False(t, u.Min)
	require.Equal(t, int64(11), u.AccessHash, "access hash is kept")
	require.Equal(t, "Full", u.FirstName, "name is kept")
	require.Equal(t, "123", u.Phone)
	require.Equal(t, fullPhoto, u.Photo, "photo is kept without ApplyMinPhoto")
	require.Equal(t, &tg.UserStatusRecently{}, u.Status, "known status is kept")
	require.True(t, u.Premium, "flags are taken from min")
	require.Equal(t, &tg.EmojiStatus{DocumentID: 7}, u.EmojiStatus)

	require.Equal(t, minPhoto, ent.users[2].Photo, "ApplyMinPhoto takes min photo")
	require.Equal(t, int64(22), ent.users[2].AccessHash)

	c := ent.channels[10]
	require.False(t, c.Min)
	require.Equal(t, int64(100), c.AccessHash)
	require.True(t, c.Creator, "participant status is kept")
	require.Equal(t, 5, c.ParticipantsCount)
	require.Equal(t, "New", c.Title, "title is taken from min")
	require.True(t, c.Verified)
}

func TestEntitiesOffsetPeer(t *testing.T) {
	ent := NewEntities(map[int64]*tg.User{
		1: {ID: 1, AccessHash: 11},
		2: {ID: 2, Min: true, AccessHash: 22},
	}, map[int64]*tg.Chat{}, map[int64]*tg.Channel{
		10: {ID: 10, Min: true, AccessHash: 100},
	})

	for _, tt := range []struct {
		peer tg.PeerClass
		want tg.InputPeerClass
	}{
		{&tg.PeerUser{UserID: 1}, &tg.InputPeerUser{UserID: 1, AccessHash: 11}},
		{&tg.PeerUser{UserID: 2}, &tg.InputPeerUser{UserID: 2, AccessHash: 22}},
		{&tg.PeerChannel{ChannelID: 10}, &tg.InputPeerChannel{ChannelID: 10, AccessHash: 100}},
		// Unknown peers get zero access hash, as in TDLib.
		{&tg.PeerUser{UserID: 3}, &tg.InputPeerUser{UserID: 3}},
		{&tg.PeerChannel{ChannelID: 30}, &tg.InputPeerChannel{ChannelID: 30}},
		{&tg.PeerChat{ChatID: 40}, &tg.InputPeerChat{ChatID: 40}},
	} {
		require.Equal(t, tt.want, ent.OffsetPeer(tt.peer), "%v", tt.peer)
	}
}
