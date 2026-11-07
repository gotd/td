package telegram

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestClientDownloadScheduler(t *testing.T) {
	ctx := context.Background()
	s := downloader.NewScheduler(downloader.SchedulerOptions{Min: 1, Start: 8, Max: 16})
	client := NewClient(TestAppID, TestAppHash, Options{DownloadScheduler: s})

	isGetFile := func(b bin.Encoder) {
		_, ok := b.(*tg.UploadGetFileRequest)
		require.True(t, ok)
	}
	data := []byte("data from pool")
	pool := tgmock.NewRequire(t)
	pool.ExpectFunc(isGetFile).ThenFlood(0)
	pool.ExpectFunc(isGetFile).ThenResult(&tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: data})

	output := new(bytes.Buffer)
	_, err := client.DownloadVia(pool, &tg.InputDocumentFileLocation{ID: 1}).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	// Flood wait of the download is seen by the client-level Scheduler.
	snap := s.Snapshot()
	require.EqualValues(t, 1, snap.Decreases)
	require.Equal(t, 4, snap.Window)
	require.Zero(t, snap.InFlight)
}
