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
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      10,
	}).ThenResult(full(main))
	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		FolderID:   1,
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      10,
	}).ThenResult(full(archive))

	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).ForEachFolder(ctx, 10, func(ctx context.Context, folder int, e Elem) error {
		got = append(got, folder)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []int{0, 0, 1}, got)
}
