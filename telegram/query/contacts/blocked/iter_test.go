package blocked

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func generateBlocked(count int) []tg.PeerBlocked {
	r := make([]tg.PeerBlocked, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, tg.PeerBlocked{
			PeerID: &tg.PeerUser{
				UserID: i + 1,
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
	mock := rpcmock.NewMock(t, require.New(t))
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
		mock.Equal(expected[i], iter.Value().Contact)
		i++
	}
	mock.NoError(iter.Err())
	mock.Equal(totalRecords, i)

	total, err := iter.Total(ctx)
	mock.NoError(err)
	mock.Equal(totalRecords, total)

	mock.ExpectCall(&tg.ContactsGetBlockedRequest{
		Offset: 0,
		Limit:  1,
	}).ThenResult(result(expected[:0], totalRecords))
	total, err = iter.FetchTotal(ctx)
	mock.NoError(err)
	mock.Equal(totalRecords, total)
}
