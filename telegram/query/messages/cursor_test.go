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
	// A short page is not the end: the count is not reached, so the iterator
	// asks for more and stops on the empty page.
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:       peer,
		Limit:      10,
		MinID:      5,
		MaxID:      50,
		OffsetID:   6,
		OffsetDate: 60,
	}).ThenResult(historyMessages())

	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).GetHistory(peer).
		MinID(5).MaxID(50).BatchSize(10).
		ForEach(ctx, func(ctx context.Context, e Elem) error {
			got = append(got, e.Msg.GetID())
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, []int{9, 8, 7, 6}, got)
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
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:       peer,
		Limit:      5,
		OffsetID:   16,
		OffsetDate: 160,
	}).ThenResult(historyMessages())
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

func TestIteratorGlobalMinOffsetPeer(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	batch := func(chatID int64, minChat bool, ids ...int) *tg.MessagesMessagesSlice {
		r := &tg.MessagesMessagesSlice{Count: 100, NextRate: ids[len(ids)-1] * 10}
		for _, id := range ids {
			r.Messages = append(r.Messages, &tg.Message{
				ID: id, Date: id * 10, PeerID: &tg.PeerChannel{ChannelID: chatID},
			})
		}
		r.Chats = []tg.ChatClass{&tg.Channel{
			ID: chatID, AccessHash: chatID * 10, Min: minChat, Photo: &tg.ChatPhotoEmpty{},
		}}
		return r
	}
	req := func(rate, id int, p tg.InputPeerClass) *tg.MessagesSearchGlobalRequest {
		return &tg.MessagesSearchGlobalRequest{
			Q: "q", Filter: &tg.InputMessagesFilterEmpty{},
			OffsetRate: rate, OffsetID: id, OffsetPeer: p, Limit: 2,
		}
	}
	mock.ExpectCall(req(0, 0, &tg.InputPeerEmpty{})).ThenResult(batch(1, false, 40, 30))
	// Last message of the second batch is in a min channel: offset_peer must
	// follow it, not stay at channel 1 from the first batch.
	mock.ExpectCall(req(300, 30, &tg.InputPeerChannel{ChannelID: 1, AccessHash: 10})).
		ThenResult(batch(2, true, 20, 10))
	mock.ExpectCall(req(100, 10, &tg.InputPeerChannel{ChannelID: 2, AccessHash: 20})).
		ThenResult(&tg.MessagesMessagesSlice{Count: 100})

	iter := NewQueryBuilder(tg.NewClient(mock)).SearchGlobal().Q("q").BatchSize(2).Iter()
	var n int
	for iter.Next(ctx) {
		n++
		if iter.Value().Msg.GetID() == 20 {
			require.Equal(t, &tg.InputPeerEmpty{}, iter.Value().Peer, "min peer is not usable")
			require.Equal(t, &tg.InputPeerChannel{ChannelID: 2, AccessHash: 20}, iter.Cursor().OffsetPeer)
		}
	}
	require.NoError(t, iter.Err())
	require.Equal(t, 4, n)
}

func TestIteratorCursorGlobalRate(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	res := &tg.MessagesMessagesSlice{Count: 100, NextRate: 555}
	for _, id := range []int{30, 20, 10} {
		res.Messages = append(res.Messages, &tg.Message{ID: id, Date: id * 10, PeerID: &tg.PeerChannel{ChannelID: 1}})
	}
	res.Chats = []tg.ChatClass{&tg.Channel{ID: 1, AccessHash: 2, Photo: &tg.ChatPhotoEmpty{}}}
	mock.ExpectCall(&tg.MessagesSearchGlobalRequest{
		Q:          "q",
		Filter:     &tg.InputMessagesFilterEmpty{},
		OffsetPeer: &tg.InputPeerEmpty{},
		OffsetRate: 999,
		Limit:      3,
	}).ThenResult(res)

	iter := NewQueryBuilder(tg.NewClient(mock)).SearchGlobal().Q("q").BatchSize(3).Iter().OffsetRate(999)
	require.True(t, iter.Next(ctx))
	require.True(t, iter.Next(ctx))
	// Mid-batch: the rate the batch was requested with (999) would restart
	// at the batch start, so the message date is used, as tdlib does.
	require.Equal(t, 200, iter.Cursor().OffsetRate)
	require.True(t, iter.Next(ctx))
	// Last element: same offset as the next request.
	require.Equal(t, 555, iter.Cursor().OffsetRate)
}
