package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestGetMediaGroup(t *testing.T) {
	ctx := context.Background()

	// Album with grouped_id 100 spans messages 5, 6, 7; message 8 belongs to a
	// different album, message 4 is a standalone message.
	const groupedID int64 = 100
	peerID := &tg.PeerUser{UserID: 1}
	album := []tg.MessageClass{
		&tg.Message{ID: 4, PeerID: peerID},
		&tg.Message{ID: 5, PeerID: peerID, GroupedID: groupedID},
		&tg.Message{ID: 6, PeerID: peerID, GroupedID: groupedID},
		&tg.Message{ID: 7, PeerID: peerID, GroupedID: groupedID},
		&tg.Message{ID: 8, PeerID: peerID, GroupedID: 200},
		&tg.MessageEmpty{ID: 9},
	}

	t.Run("User", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		mock.ExpectFunc(func(b bin.Encoder) {
			_, ok := b.(*tg.MessagesGetMessagesRequest)
			require.True(t, ok)
		}).ThenResult(&tg.MessagesMessages{Messages: album})

		group, err := NewQueryBuilder(raw).GetMediaGroup(ctx, &tg.InputPeerUser{UserID: 1}, 6)
		require.NoError(t, err)
		require.Len(t, group, 3)
		require.Equal(t, 5, group[0].ID)
		require.Equal(t, 6, group[1].ID)
		require.Equal(t, 7, group[2].ID)
	})

	t.Run("Channel", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		mock.ExpectFunc(func(b bin.Encoder) {
			req, ok := b.(*tg.ChannelsGetMessagesRequest)
			require.True(t, ok)
			require.Equal(t, &tg.InputChannel{ChannelID: 10, AccessHash: 20}, req.Channel)
		}).ThenResult(&tg.MessagesChannelMessages{Messages: album})

		group, err := NewQueryBuilder(raw).GetMediaGroup(ctx,
			&tg.InputPeerChannel{ChannelID: 10, AccessHash: 20}, 6)
		require.NoError(t, err)
		require.Len(t, group, 3)
	})

	t.Run("NotInGroup", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		mock.ExpectFunc(func(b bin.Encoder) {}).
			ThenResult(&tg.MessagesMessages{Messages: album})

		_, err := NewQueryBuilder(raw).GetMediaGroup(ctx, &tg.InputPeerUser{UserID: 1}, 4)
		require.Error(t, err)
	})
}
