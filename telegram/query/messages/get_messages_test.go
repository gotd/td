package messages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestGetMessages(t *testing.T) {
	ctx := context.Background()

	peerID := &tg.PeerUser{UserID: 1}
	result := []tg.MessageClass{
		&tg.Message{ID: 10, PeerID: peerID},
		&tg.MessageEmpty{ID: 11},
		&tg.Message{ID: 12, PeerID: peerID},
	}

	t.Run("User", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		mock.ExpectFunc(func(b bin.Encoder) {
			req, ok := b.(*tg.MessagesGetMessagesRequest)
			require.True(t, ok)
			require.Equal(t, []tg.InputMessageClass{
				&tg.InputMessageID{ID: 10},
				&tg.InputMessageID{ID: 11},
				&tg.InputMessageID{ID: 12},
			}, req.ID)
		}).ThenResult(&tg.MessagesMessages{Messages: result})

		msgs, err := NewQueryBuilder(raw).GetMessages(ctx, &tg.InputPeerUser{UserID: 1}, 10, 11, 12)
		require.NoError(t, err)
		require.Equal(t, result, msgs)
	})

	t.Run("Channel", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		mock.ExpectFunc(func(b bin.Encoder) {
			req, ok := b.(*tg.ChannelsGetMessagesRequest)
			require.True(t, ok)
			require.Equal(t, &tg.InputChannel{ChannelID: 10, AccessHash: 20}, req.Channel)
			require.Equal(t, []tg.InputMessageClass{&tg.InputMessageID{ID: 42}}, req.ID)
		}).ThenResult(&tg.MessagesChannelMessages{Messages: []tg.MessageClass{
			&tg.Message{ID: 42, PeerID: &tg.PeerChannel{ChannelID: 10}},
		}})

		msgs, err := NewQueryBuilder(raw).GetMessages(ctx,
			&tg.InputPeerChannel{ChannelID: 10, AccessHash: 20}, 42)
		require.NoError(t, err)
		require.Len(t, msgs, 1)
		require.Equal(t, 42, msgs[0].(*tg.Message).ID)
	})

	t.Run("NoIDs", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		raw := tg.NewClient(mock)
		_, err := NewQueryBuilder(raw).GetMessages(ctx, &tg.InputPeerUser{UserID: 1})
		require.Error(t, err)
	})
}
