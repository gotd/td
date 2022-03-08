package members

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/tg"
)

func TestChannelMembers_Count(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestChannel())
	members, err := Channel(ctx, ch)
	a.NoError(err)

	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch.InputChannel(),
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  0,
		Limit:   1,
	}).ThenErr(testutil.TestError())
	_, err = members.Count(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch.InputChannel(),
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  0,
		Limit:   1,
	}).ThenResult(&tg.ChannelsChannelParticipantsNotModified{})
	_, err = members.Count(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch.InputChannel(),
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  0,
		Limit:   1,
	}).ThenResult(&tg.ChannelsChannelParticipants{
		Count: 10,
	})
	count, err := members.Count(ctx)
	a.NoError(err)
	a.Equal(10, count)
}
