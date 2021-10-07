package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func generateDialogs(count int) []tg.DialogClass {
	r := make([]tg.DialogClass, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, &tg.Dialog{
			Peer: &tg.PeerChannel{ChannelID: int64(i)},
		})
	}

	return r
}

func result(r []tg.DialogClass, count int) tg.MessagesDialogsClass {
	msgs := make([]tg.MessageClass, 0, len(r))
	for i, dlg := range r {
		msgs = append(msgs, &tg.Message{
			ID:     i,
			PeerID: dlg.GetPeer(),
		})
	}

	chats := make([]tg.ChatClass, 0, len(r))
	for i, dlg := range r {
		id := dlg.GetPeer().(*tg.PeerChannel).ChannelID
		chats = append(chats, &tg.Channel{
			ID:         id,
			AccessHash: 10,
			Photo: &tg.ChatPhoto{
				PhotoID: int64(i),
			},
		})
	}

	return &tg.MessagesDialogsSlice{
		Dialogs:  r,
		Messages: msgs,
		Chats:    chats,
		Count:    count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	limit := 10
	totalRows := 3 * limit
	expected := generateDialogs(totalRows)
	raw := tg.NewClient(mock)

	mock.Expect().ThenResult(result(expected[0:limit], totalRows))
	mock.Expect().ThenResult(result(expected[limit:2*limit], totalRows))
	mock.Expect().ThenResult(result(expected[2*limit:3*limit], totalRows))
	mock.Expect().ThenResult(result(expected[3*limit:], totalRows))

	iter := NewQueryBuilder(raw).GetDialogs().BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i].GetPeer(), iter.Value().Dialog.GetPeer())
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, totalRows, i)

	total, err := iter.Total(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRows, total)

	mock.ExpectCall(&tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      1,
	}).ThenResult(result(expected[:0], totalRows))
	total, err = iter.FetchTotal(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRows, total)
}
