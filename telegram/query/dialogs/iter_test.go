package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
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
			PeerID: dlg.(*tg.Dialog).Peer,
		})
	}

	chats := make([]tg.ChatClass, 0, len(r))
	for i, dlg := range r {
		id := dlg.(*tg.Dialog).Peer.(*tg.PeerChannel).ChannelID
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
		require.Equal(t, expected[i].(*tg.Dialog).Peer, iter.Value().Dialog.(*tg.Dialog).Peer)
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

func TestIteratorPeerlessDialog(t *testing.T) {
	ctx := context.Background()
	community := &tg.DialogCommunity{}
	first := &tg.Message{ID: 12, Date: 34, PeerID: &tg.PeerUser{UserID: 8}}
	last := &tg.Message{ID: 21, Date: 43, PeerID: &tg.PeerUser{UserID: 9}}
	dialogs := []tg.DialogClass{
		&tg.Dialog{Peer: first.PeerID, TopMessage: first.ID},
		community,
		&tg.Dialog{Peer: last.PeerID, TopMessage: last.ID},
	}
	messages := []tg.MessageClass{first, last}
	users := []tg.UserClass{
		&tg.User{ID: 8, AccessHash: 56},
		&tg.User{ID: 9, AccessHash: 65},
	}

	for _, test := range []struct {
		name     string
		response tg.MessagesDialogsClass
		paginate bool
	}{
		{"Complete", &tg.MessagesDialogs{Dialogs: dialogs, Messages: messages, Users: users}, false},
		{"Slice", &tg.MessagesDialogsSlice{Count: 3, Dialogs: dialogs, Messages: messages, Users: users}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			mock := tgmock.NewRequire(t)
			mock.ExpectCall(&tg.MessagesGetDialogsRequest{OffsetPeer: &tg.InputPeerEmpty{}, Limit: 100}).
				ThenResult(test.response)
			if test.paginate {
				mock.ExpectCall(&tg.MessagesGetDialogsRequest{
					OffsetID: last.ID, OffsetDate: last.Date,
					OffsetPeer: &tg.InputPeerUser{UserID: 9, AccessHash: 65}, Limit: 100,
				}).ThenResult(&tg.MessagesDialogsSlice{})
			}

			iter := NewIterator(NewQueryBuilder(tg.NewClient(mock)).GetDialogs(), 100)
			require.True(t, iter.Next(ctx))
			require.Equal(t, first, iter.Value().Last)
			require.True(t, iter.Next(ctx))
			require.Equal(t, community, iter.Value().Dialog)
			require.IsType(t, &tg.InputPeerEmpty{}, iter.Value().Peer)
			require.Nil(t, iter.Value().Last)
			require.True(t, iter.Next(ctx))
			require.Equal(t, last, iter.Value().Last)
			total, err := iter.Total(ctx)
			require.NoError(t, err)
			require.Equal(t, len(dialogs), total)
			if test.paginate {
				require.False(t, iter.Next(ctx))
				require.NoError(t, iter.Err())
			}
		})
	}
}
