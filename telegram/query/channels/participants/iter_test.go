package participants

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func generateParticipants(count int) []tg.ChannelParticipantClass {
	r := make([]tg.ChannelParticipantClass, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, &tg.ChannelParticipant{
			UserID: int64(i),
			Date:   i,
		})
	}

	return r
}

func result(r []tg.ChannelParticipantClass, count int) tg.ChannelsChannelParticipantsClass {
	return &tg.ChannelsChannelParticipants{
		Participants: r,
		Count:        count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	limit := 10
	totalRecords := 3 * limit
	expected := generateParticipants(totalRecords)
	raw := tg.NewClient(mock)
	ch := &tg.InputChannel{
		ChannelID:  10,
		AccessHash: 10,
	}

	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  0,
		Limit:   limit,
	}).ThenResult(result(expected[0:limit], totalRecords))
	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  limit,
		Limit:   limit,
	}).ThenResult(result(expected[limit:2*limit], totalRecords))
	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  2 * limit,
		Limit:   limit,
	}).ThenResult(result(expected[2*limit:3*limit], totalRecords))
	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  3 * limit,
		Limit:   limit,
	}).ThenResult(result(expected[3*limit:], totalRecords))

	iter := NewQueryBuilder(raw).GetParticipants(ch).BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i], iter.Value().Participant)
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, totalRecords, i)

	total, err := iter.Total(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)

	mock.ExpectCall(&tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsRecent{},
		Offset:  0,
		Limit:   1,
	}).ThenResult(result(expected[:0], totalRecords))
	total, err = iter.FetchTotal(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)
}
