package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// RecommendedChannels is the result of a channel recommendations query.
//
// channels.getChannelRecommendations has no offset or hash parameter, so it
// can't be paginated: the server returns the available recommendations in a
// single response. Non-Premium accounts receive a server-capped subset of the
// recommendations, while Count still reports the full total — so a caller can
// detect that more recommendations exist (e.g. to prompt for Premium) by
// comparing len(Channels) with Count.
//
// See https://core.telegram.org/method/channels.getChannelRecommendations.
type RecommendedChannels struct {
	// Channels is the list of recommended channels returned by the server.
	Channels []Channel
	// Count is the total number of recommendations available, which may be
	// greater than len(Channels) for non-Premium accounts.
	Count int
}

// RecommendedChannels returns public channels recommended based on similarities
// in the subscriber bases of this channel and others.
func (c Channel) RecommendedChannels(ctx context.Context) (RecommendedChannels, error) {
	req := &tg.ChannelsGetChannelRecommendationsRequest{}
	req.SetChannel(c.InputChannel())
	return c.m.recommendedChannels(ctx, req)
}

// RecommendedChannels returns channels recommended for the current user, based
// on the channels the user has joined.
func (m *Manager) RecommendedChannels(ctx context.Context) (RecommendedChannels, error) {
	return m.recommendedChannels(ctx, &tg.ChannelsGetChannelRecommendationsRequest{})
}

func (m *Manager) recommendedChannels(
	ctx context.Context,
	req *tg.ChannelsGetChannelRecommendationsRequest,
) (RecommendedChannels, error) {
	res, err := m.api.ChannelsGetChannelRecommendations(ctx, req)
	if err != nil {
		return RecommendedChannels{}, errors.Wrap(err, "get channel recommendations")
	}

	chats := res.GetChats()
	if err := m.applyChats(ctx, chats...); err != nil {
		return RecommendedChannels{}, errors.Wrap(err, "apply chats")
	}

	result := RecommendedChannels{
		Channels: make([]Channel, 0, len(chats)),
	}
	for _, chat := range chats {
		ch, ok := chat.(*tg.Channel)
		if !ok {
			continue
		}
		result.Channels = append(result.Channels, m.Channel(ch))
	}

	// messages.chatsSlice carries the total count; messages.chats does not, in
	// which case the full set was returned and the count equals its length.
	if slice, ok := res.(*tg.MessagesChatsSlice); ok {
		result.Count = slice.Count
	} else {
		result.Count = len(result.Channels)
	}

	return result, nil
}
