package featured

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func generateStickers(count int) []tg.StickerSetCoveredClass {
	r := make([]tg.StickerSetCoveredClass, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, &tg.StickerSetCovered{
			Set: tg.StickerSet{
				ID:         int64(i + 1),
				AccessHash: int64(i + 1),
			},
			Cover: &tg.Document{
				ID:            int64(i + 1),
				AccessHash:    int64(i + 1),
				FileReference: []uint8{uint8(i)},
			},
		})
	}

	return r
}

func result(r []tg.StickerSetCoveredClass, count int) tg.MessagesFeaturedStickersClass {
	return &tg.MessagesFeaturedStickers{
		Sets:  r,
		Count: count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	limit := 10
	totalRecords := 3 * limit
	expected := generateStickers(totalRecords)
	raw := tg.NewClient(mock)

	mock.ExpectCall(&tg.MessagesGetOldFeaturedStickersRequest{
		Offset: 0,
		Limit:  limit,
	}).ThenResult(result(expected[0:limit], totalRecords))
	mock.ExpectCall(&tg.MessagesGetOldFeaturedStickersRequest{
		Offset: limit,
		Limit:  limit,
	}).ThenResult(result(expected[limit:2*limit], totalRecords))
	mock.ExpectCall(&tg.MessagesGetOldFeaturedStickersRequest{
		Offset: 2 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[2*limit:3*limit], totalRecords))
	mock.ExpectCall(&tg.MessagesGetOldFeaturedStickersRequest{
		Offset: 3 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[3*limit:], totalRecords))

	iter := NewQueryBuilder(raw).GetOldFeaturedStickers().BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i], iter.Value().Sticker)
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, totalRecords, i)

	total, err := iter.Total(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)

	mock.ExpectCall(&tg.MessagesGetOldFeaturedStickersRequest{
		Offset: 0,
		Limit:  1,
	}).ThenResult(result(expected[:0], totalRecords))
	total, err = iter.FetchTotal(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)
}
