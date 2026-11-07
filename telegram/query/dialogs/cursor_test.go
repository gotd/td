package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestIteratorCursor(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	dialogs := generateDialogs(5)
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      10,
	}).ThenResult(result(dialogs, 5))

	iter := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().BatchSize(10).Iter()
	require.Equal(t, Cursor{OffsetPeer: &tg.InputPeerEmpty{}}, iter.Cursor())
	for range 3 {
		require.True(t, iter.Next(ctx))
	}
	c := iter.Cursor()
	require.Equal(t, 2, c.OffsetID)
	require.Equal(t, &tg.InputPeerChannel{ChannelID: 2, AccessHash: 10}, c.OffsetPeer)

	resumed := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().Iter().Resume(Cursor{})
	require.Equal(t, Cursor{OffsetPeer: &tg.InputPeerEmpty{}}, resumed.Cursor())
}

func TestForEachFolder(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	main, archive := generateDialogs(2), generateDialogs(3)[2:]
	full := func(r []tg.DialogClass) tg.MessagesDialogsClass {
		slice := result(r, len(r)).(*tg.MessagesDialogsSlice)
		return &tg.MessagesDialogs{Dialogs: slice.Dialogs, Messages: slice.Messages, Chats: slice.Chats}
	}
	mainRes := full(main).(*tg.MessagesDialogs)
	// Main list includes the archive folder entry, which must be skipped.
	mainRes.Dialogs = append([]tg.DialogClass{&tg.DialogFolder{
		Folder: tg.Folder{ID: 1, Title: "Archived"},
		Peer:   &tg.PeerChannel{ChannelID: 2},
		Pinned: true,
	}}, mainRes.Dialogs...)

	// folder_id must be sent for the main list too: without it the server
	// returns dialogs of all folders.
	mainReq := &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      10,
	}
	mainReq.SetFolderID(0)
	archiveReq := &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      10,
	}
	archiveReq.SetFolderID(1)
	mock.ExpectCall(mainReq).ThenResult(mainRes)
	mock.ExpectCall(archiveReq).ThenResult(full(archive))

	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).ForEachFolder(ctx, 10, func(ctx context.Context, folder int, e Elem) error {
		got = append(got, folder)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []int{0, 0, 1}, got)
}

func TestIteratorMinOffsetPeer(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	res := result(generateDialogs(3), 10).(*tg.MessagesDialogsSlice)
	// The last dialog's channel is known only as min: its access hash is
	// still enough for offset_peer, so iteration must not fail.
	res.Chats[2].(*tg.Channel).Min = true
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      3,
	}).ThenResult(res)

	iter := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().BatchSize(3).Iter()
	for range 3 {
		require.True(t, iter.Next(ctx))
	}
	require.NoError(t, iter.Err())
	require.Equal(t, &tg.InputPeerChannel{ChannelID: 2, AccessHash: 10}, iter.offsetPeer)
	require.Equal(t, &tg.InputPeerEmpty{}, iter.Value().Peer, "min peer is not usable for requests")
}

func TestIteratorUnknownOffsetPeer(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	res := result(generateDialogs(3), 10).(*tg.MessagesDialogsSlice)
	// The last dialog's channel is not returned at all: like TDLib, offset_peer
	// is built from its ID with zero access hash, so iteration must not fail.
	res.Chats = res.Chats[:2]
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      3,
	}).ThenResult(res)

	iter := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().BatchSize(3).Iter()
	for range 3 {
		require.True(t, iter.Next(ctx))
	}
	require.NoError(t, iter.Err())
	require.Equal(t, &tg.InputPeerChannel{ChannelID: 2}, iter.offsetPeer)
}

func TestIteratorCursorPinned(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	dialogs := generateDialogs(3)
	res := result(dialogs, 10).(*tg.MessagesDialogsSlice)
	// First dialog is pinned with an old last message.
	dialogs[0].(*tg.Dialog).Pinned = true
	msgs := res.Messages
	msgs[0].(*tg.Message).Date = 1
	msgs[1].(*tg.Message).Date = 300
	msgs[2].(*tg.Message).Date = 200
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      3,
	}).ThenResult(res)

	iter := NewQueryBuilder(tg.NewClient(mock)).GetDialogs().BatchSize(3).Iter()
	require.True(t, iter.Next(ctx))
	// Resuming from the pinned dialog date would skip newer dialogs.
	require.Equal(t, Cursor{OffsetPeer: &tg.InputPeerEmpty{}}, iter.Cursor())

	require.True(t, iter.Next(ctx))
	require.Equal(t, Cursor{
		OffsetID:   1,
		OffsetDate: 300,
		OffsetPeer: &tg.InputPeerChannel{ChannelID: 1, AccessHash: 10},
	}, iter.Cursor())
}
