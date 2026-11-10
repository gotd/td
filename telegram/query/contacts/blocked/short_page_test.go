package blocked

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

// TestIteratorShortPage checks that a short non-empty page mid-stream does not
// end iteration: it continues from the new offset until the count is reached.
func TestIteratorShortPage(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	const (
		limit = 10
		total = 30
	)
	expected := generateBlocked(total)
	raw := tg.NewClient(mock)

	// Pages of 10, 4 (short), 10, 6.
	for _, p := range [][2]int{{0, 10}, {10, 14}, {14, 24}, {24, 30}} {
		mock.ExpectCall(&tg.ContactsGetBlockedRequest{
			Offset: p[0],
			Limit:  limit,
		}).ThenResult(result(expected[p[0]:p[1]], total))
	}

	iter := NewQueryBuilder(raw).GetBlocked().BatchSize(limit).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i], iter.Value().Contact)
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, total, i)
}

// TestIteratorEmptyPageBeforeCount checks that an empty page ends iteration
// even if the server-reported count is not reached.
func TestIteratorEmptyPageBeforeCount(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	const (
		limit = 10
		total = 50
	)
	expected := generateBlocked(15)
	raw := tg.NewClient(mock)

	for _, p := range [][2]int{{0, 10}, {10, 15}, {15, 15}} {
		mock.ExpectCall(&tg.ContactsGetBlockedRequest{
			Offset: p[0],
			Limit:  limit,
		}).ThenResult(result(expected[p[0]:p[1]], total))
	}

	iter := NewQueryBuilder(raw).GetBlocked().BatchSize(limit).Iter()
	i := 0
	for iter.Next(ctx) {
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, len(expected), i)
	// The last batch is consumed: no more requests.
	require.False(t, iter.Next(ctx))
}
