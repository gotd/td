package photos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func generatePhotos(count int) []tg.PhotoClass {
	r := make([]tg.PhotoClass, 0, count)

	for i := 0; i < count; i++ {
		r = append(r, &tg.Photo{
			ID:            int64(i + 1),
			AccessHash:    int64(i + 1),
			FileReference: []uint8{uint8(i)},
		})
	}

	return r
}

func result(r []tg.PhotoClass, count int) tg.PhotosPhotosClass {
	return &tg.PhotosPhotosSlice{
		Photos: r,
		Count:  count,
	}
}

func TestIterator(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	limit := 10
	totalRecords := 3 * limit
	expected := generatePhotos(totalRecords)
	raw := tg.NewClient(mock)

	mock.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Offset: 0,
		Limit:  limit,
	}).ThenResult(result(expected[0:limit], totalRecords))
	mock.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Offset: limit,
		Limit:  limit,
	}).ThenResult(result(expected[limit:2*limit], totalRecords))
	mock.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Offset: 2 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[2*limit:3*limit], totalRecords))
	mock.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Offset: 3 * limit,
		Limit:  limit,
	}).ThenResult(result(expected[3*limit:], totalRecords))

	iter := NewQueryBuilder(raw).GetUserPhotos(&tg.InputUserSelf{}).BatchSize(10).Iter()
	i := 0
	for iter.Next(ctx) {
		require.Equal(t, expected[i], iter.Value().Photo)
		i++
	}
	require.NoError(t, iter.Err())
	require.Equal(t, totalRecords, i)

	total, err := iter.Total(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)

	mock.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Offset: 0,
		Limit:  1,
	}).ThenResult(result(expected[:0], totalRecords))
	total, err = iter.FetchTotal(ctx)
	require.NoError(t, err)
	require.Equal(t, totalRecords, total)
}
