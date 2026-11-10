package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

// dialogsPage returns dialogs with channel IDs ids, whose top message ID and
// date are the channel ID.
func dialogsPage(count int, ids ...int) *tg.MessagesDialogsSlice {
	r := &tg.MessagesDialogsSlice{Count: count}
	for _, id := range ids {
		p := &tg.PeerChannel{ChannelID: int64(id)}
		r.Dialogs = append(r.Dialogs, &tg.Dialog{Peer: p, TopMessage: id})
		r.Messages = append(r.Messages, &tg.Message{ID: id, Date: id, PeerID: p})
		r.Chats = append(r.Chats, &tg.Channel{ID: int64(id), AccessHash: 10, Photo: &tg.ChatPhotoEmpty{}})
	}
	return r
}

func dialogIDs(from, to int) []int {
	var r []int
	for id := from; id >= to; id-- {
		r = append(r, id)
	}
	return r
}

func dialogsRequest(offsetID int) *tg.MessagesGetDialogsRequest {
	req := &tg.MessagesGetDialogsRequest{
		Limit:      10,
		OffsetID:   offsetID,
		OffsetDate: offsetID,
		OffsetPeer: &tg.InputPeerEmpty{},
	}
	if offsetID != 0 {
		req.OffsetPeer = &tg.InputPeerChannel{ChannelID: int64(offsetID), AccessHash: 10}
	}
	return req
}

func collectDialogs(t *testing.T, mock *tgmock.Mock) []int {
	t.Helper()
	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().BatchSize(10).
		ForEach(context.Background(), func(ctx context.Context, e Elem) error {
			got = append(got, e.Last.GetID())
			return nil
		})
	require.NoError(t, err)
	return got
}

func TestIteratorShortPage(t *testing.T) {
	mock := tgmock.NewRequire(t)

	// Pages of 10, 4 (short), 10, 6: count is reached without an empty page.
	mock.ExpectCall(dialogsRequest(0)).ThenResult(dialogsPage(30, dialogIDs(30, 21)...))
	mock.ExpectCall(dialogsRequest(21)).ThenResult(dialogsPage(30, dialogIDs(20, 17)...))
	mock.ExpectCall(dialogsRequest(17)).ThenResult(dialogsPage(30, dialogIDs(16, 7)...))
	mock.ExpectCall(dialogsRequest(7)).ThenResult(dialogsPage(30, dialogIDs(6, 1)...))

	require.Equal(t, dialogIDs(30, 1), collectDialogs(t, mock))
}

func TestIteratorEmptyPageBeforeCount(t *testing.T) {
	mock := tgmock.NewRequire(t)

	mock.ExpectCall(dialogsRequest(0)).ThenResult(dialogsPage(50, dialogIDs(30, 21)...))
	mock.ExpectCall(dialogsRequest(21)).ThenResult(dialogsPage(50, dialogIDs(20, 17)...))
	mock.ExpectCall(dialogsRequest(17)).ThenResult(dialogsPage(50))

	require.Equal(t, dialogIDs(30, 17), collectDialogs(t, mock))
}

func TestIteratorOffsetNotAdvancing(t *testing.T) {
	mock := tgmock.NewRequire(t)

	mock.ExpectCall(dialogsRequest(0)).ThenResult(dialogsPage(100, dialogIDs(30, 28)...))
	// Server ignores the offset and returns the same page again.
	mock.ExpectCall(dialogsRequest(28)).ThenResult(dialogsPage(100, dialogIDs(30, 28)...))

	require.Equal(t, dialogIDs(30, 28), collectDialogs(t, mock))
}
