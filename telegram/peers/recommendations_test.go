package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func getTestRecommendedChannel() *tg.Channel {
	u := &tg.Channel{
		Broadcast:  true,
		ID:         12,
		AccessHash: 12,
		Title:      "Recommended",
		Username:   "recommended",
		Photo:      &tg.ChatPhotoEmpty{},
	}
	u.SetFlags()
	return u
}

func TestChannel_RecommendedChannels(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())
	rec := getTestRecommendedChannel()

	req := &tg.ChannelsGetChannelRecommendationsRequest{}
	req.SetChannel(ch.InputChannel())

	// Error path.
	mock.ExpectCall(req).ThenRPCErr(getTestError())
	_, err := ch.RecommendedChannels(ctx)
	a.Error(err)

	// messages.chats: full set, Count equals number of returned channels.
	mock.ExpectCall(req).ThenResult(&tg.MessagesChats{
		Chats: []tg.ChatClass{rec},
	})
	got, err := ch.RecommendedChannels(ctx)
	a.NoError(err)
	a.Len(got.Channels, 1)
	a.Equal(1, got.Count)
	a.Equal(rec.ID, got.Channels[0].ID())
	username, ok := got.Channels[0].Username()
	a.True(ok)
	a.Equal("recommended", username)

	// messages.chatsSlice: capped subset, Count reports the true total.
	mock.ExpectCall(req).ThenResult(&tg.MessagesChatsSlice{
		Count: 82,
		Chats: []tg.ChatClass{rec},
	})
	got, err = ch.RecommendedChannels(ctx)
	a.NoError(err)
	a.Len(got.Channels, 1)
	a.Equal(82, got.Count, "Count must report the total, not the returned length")
}

func TestManager_RecommendedChannels(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	rec := getTestRecommendedChannel()

	// No channel set: recommendations for the current user.
	req := &tg.ChannelsGetChannelRecommendationsRequest{}

	mock.ExpectCall(req).ThenResult(&tg.MessagesChatsSlice{
		Count: 5,
		Chats: []tg.ChatClass{rec, &tg.ChatEmpty{ID: 999}},
	})
	got, err := m.RecommendedChannels(ctx)
	a.NoError(err)
	// Non-channel chats are skipped, but the total Count is preserved.
	a.Len(got.Channels, 1)
	a.Equal(5, got.Count)
	a.Equal(rec.ID, got.Channels[0].ID())
}
