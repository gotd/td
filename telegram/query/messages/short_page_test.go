package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

// historyPage returns channel messages with given IDs and count.
func historyPage(count int, ids ...int) *tg.MessagesChannelMessages {
	r := &tg.MessagesChannelMessages{Count: count}
	for _, id := range ids {
		r.Messages = append(r.Messages, &tg.Message{ID: id, Date: id * 10, PeerID: &tg.PeerChannel{ChannelID: 1}})
	}
	r.Chats = []tg.ChatClass{&tg.Channel{ID: 1, AccessHash: 2, Photo: &tg.ChatPhotoEmpty{}}}
	return r
}

func ids(from, to int) []int {
	var r []int
	for id := from; id >= to; id-- {
		r = append(r, id)
	}
	return r
}

func historyRequest(peer tg.InputPeerClass, offsetID int) *tg.MessagesGetHistoryRequest {
	return &tg.MessagesGetHistoryRequest{
		Peer:       peer,
		Limit:      10,
		OffsetID:   offsetID,
		OffsetDate: offsetID * 10,
	}
}

func collectHistory(t *testing.T, mock *tgmock.Mock, peer tg.InputPeerClass) []int {
	t.Helper()
	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).GetHistory(peer).BatchSize(10).
		ForEach(context.Background(), func(ctx context.Context, e Elem) error {
			got = append(got, e.Msg.GetID())
			return nil
		})
	require.NoError(t, err)
	return got
}

func TestIteratorShortPage(t *testing.T) {
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}

	// Pages of 10, 4 (short, e.g. deleted messages), 10, 6.
	mock.ExpectCall(historyRequest(peer, 0)).ThenResult(historyPage(30, ids(30, 21)...))
	mock.ExpectCall(historyRequest(peer, 21)).ThenResult(historyPage(30, ids(20, 17)...))
	mock.ExpectCall(historyRequest(peer, 17)).ThenResult(historyPage(30, ids(16, 7)...))
	mock.ExpectCall(historyRequest(peer, 7)).ThenResult(historyPage(30, ids(6, 1)...))

	require.Equal(t, ids(30, 1), collectHistory(t, mock, peer))
}

func TestIteratorOnlyEmptyMessagesPage(t *testing.T) {
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}

	empty := &tg.MessagesChannelMessages{Count: 20}
	for _, id := range ids(20, 11) {
		empty.Messages = append(empty.Messages, &tg.MessageEmpty{ID: id})
	}
	mock.ExpectCall(historyRequest(peer, 0)).ThenResult(empty)
	// messageEmpty has no date, so the date offset stays.
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer:     peer,
		Limit:    10,
		OffsetID: 11,
	}).ThenResult(historyPage(20, ids(10, 1)...))

	require.Equal(t, ids(10, 1), collectHistory(t, mock, peer))
}

func TestIteratorInexactCount(t *testing.T) {
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}

	page := func(msgIDs ...int) *tg.MessagesMessagesSlice {
		r := &tg.MessagesMessagesSlice{Count: 5, Inexact: true}
		for _, id := range msgIDs {
			r.Messages = append(r.Messages, &tg.Message{ID: id, Date: id * 10, PeerID: &tg.PeerChannel{ChannelID: 1}})
		}
		return r
	}
	// Inexact count is below the real number of messages: ignore it.
	mock.ExpectCall(historyRequest(peer, 0)).ThenResult(page(ids(20, 11)...))
	mock.ExpectCall(historyRequest(peer, 11)).ThenResult(page(ids(10, 8)...))
	mock.ExpectCall(historyRequest(peer, 8)).ThenResult(page())

	require.Equal(t, ids(20, 8), collectHistory(t, mock, peer))
}

func TestIteratorOffsetNotAdvancing(t *testing.T) {
	mock := tgmock.NewRequire(t)
	peer := &tg.InputPeerChannel{ChannelID: 1, AccessHash: 2}

	mock.ExpectCall(historyRequest(peer, 0)).ThenResult(historyPage(100, ids(20, 18)...))
	// Server ignores the offset and returns the same page again.
	mock.ExpectCall(historyRequest(peer, 18)).ThenResult(historyPage(100, ids(20, 18)...))

	require.Equal(t, ids(20, 18), collectHistory(t, mock, peer))
}
