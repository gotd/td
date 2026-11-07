package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func historyMessages(ids ...int) *tg.MessagesChannelMessages {
	r := &tg.MessagesChannelMessages{Count: 100}
	for _, id := range ids {
		r.Messages = append(r.Messages, &tg.Message{ID: id, Date: id * 10, PeerID: &tg.PeerChannel{ChannelID: 1}})
	}
	r.Chats = []tg.ChatClass{&tg.Channel{ID: 1, AccessHash: 2, Photo: &tg.ChatPhotoEmpty{}}}
	return r
}

func TestGetHistoryMinMaxID(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: 10,
		MinID: 5,
		MaxID: 50,
	}).ThenResult(historyMessages(9, 8, 7, 6))

	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).GetHistory(peer).
		MinID(5).MaxID(50).BatchSize(10).
		ForEach(ctx, func(ctx context.Context, e Elem) error {
			got = append(got, e.Msg.GetID())
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, []int{9, 8, 7, 6}, got, "short batch is the last one")
}

func TestIteratorCursor(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}
	q := NewQueryBuilder(tg.NewClient(mock))

	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: 5,
	}).ThenResult(historyMessages(20, 19, 18, 17, 16))

	iter := q.GetHistory(peer).BatchSize(5).Iter()
	require.Equal(t, Cursor{OffsetPeer: &tg.InputPeerEmpty{}}, iter.Cursor())
	for range 3 {
		require.True(t, iter.Next(ctx))
	}
	c := iter.Cursor()
	require.Equal(t, 18, c.OffsetID)
	require.Equal(t, 180, c.OffsetDate)
	require.Equal(t, peer, c.OffsetPeer)

	// Resume in new iterator: continues after message 18.
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:       peer,
		Limit:      5,
		OffsetID:   18,
		OffsetDate: 180,
	}).ThenResult(historyMessages(17, 16))
	resumed := q.GetHistory(peer).BatchSize(5).Iter().Resume(c)
	var got []int
	for resumed.Next(ctx) {
		got = append(got, resumed.Value().Msg.GetID())
	}
	require.NoError(t, resumed.Err())
	require.Equal(t, []int{17, 16}, got)
}

func TestGetMessagesWithEntities(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}
	res := historyMessages(3)
	res.Messages = append(res.Messages, &tg.MessageEmpty{ID: 4})
	res.Users = []tg.UserClass{&tg.User{ID: 7, AccessHash: 8}}
	mock.ExpectCall(&tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{ChannelID: 1, AccessHash: 2},
		ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: 3}, &tg.InputMessageID{ID: 4}},
	}).ThenResult(res)

	elems, err := NewQueryBuilder(tg.NewClient(mock)).GetMessagesWithEntities(ctx, peer, 3, 4)
	require.NoError(t, err)
	require.Len(t, elems, 1)
	require.Equal(t, 3, elems[0].Msg.GetID())
	require.Equal(t, peer, elems[0].Peer)
	_, ok := elems[0].Entities.User(7)
	require.True(t, ok)
}
