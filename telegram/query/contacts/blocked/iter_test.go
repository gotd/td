package blocked

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func generateBlocked(count int) []tg.PeerBlocked {
	r := make([]tg.PeerBlocked, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, tg.PeerBlocked{
			PeerID: &tg.PeerUser{
				UserID: int64(i + 1),
			},
			Date: i,
		})
	}

	return r
}

func result(r []tg.PeerBlocked, count int) tg.ContactsBlockedClass {
	return &tg.ContactsBlockedSlice{
		Blocked: r,
		Count:   count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	limit := 10
	totalRecords := 3 * limit
	expected := generateBlocked(totalRecords)
	raw := tg.NewClient(mock)

	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: 0,
		Limit:  limit,
	}).ThenResult(result(expected[0:limit], totalRecords))
	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: limit,
		Limit:  limit,
	}).ThenResult(result(expected[limit:2*limit], totalRecords))
	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: 2 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[2*limit:3*limit], totalRecords))
	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: 3 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[3*limit:], totalRecords))

	iter := NewQueryBuilder(raw).GetBlocked().BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i], iter.Value().Contact)
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, totalRecords, i)

	total, err := iter.Total(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)

	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: 0,
		Limit:  1,
	}).ThenResult(result(expected[:0], totalRecords))
	total, err = iter.FetchTotal(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)
}
