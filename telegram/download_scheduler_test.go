package telegram

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

// floodOncePool returns pool that answers upload.getFile with flood wait once
// and then with data.
func floodOncePool(t *testing.T, data []byte) *tgmock.Mock {
	t.Helper()
	isGetFile := func(b bin.Encoder) {
		_, ok := b.(*tg.UploadGetFileRequest)
		require.True(t, ok)
	}
	pool := tgmock.NewRequire(t)
	pool.ExpectFunc(isGetFile).ThenFlood(0)
	pool.ExpectFunc(isGetFile).ThenResult(&tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: data})
	return pool
}

func TestClientDownloadScheduler(t *testing.T) {
	ctx := context.Background()
	opts := downloader.SchedulerOptions{Min: 1, Start: 8, Max: 16}
	client := NewClient(TestAppID, TestAppHash, Options{DownloadScheduler: &opts})
	data := []byte("data from pool")

	dc2, dc4 := client.DownloadScheduler(2), client.DownloadScheduler(4)
	require.NotNil(t, dc2)
	require.Same(t, dc2, client.DownloadScheduler(2), "one Scheduler per DC")
	require.NotSame(t, dc2, dc4)

	t.Run("ViaUsesNone", func(t *testing.T) {
		output := new(bytes.Buffer)
		_, err := client.DownloadVia(floodOncePool(t, data), &tg.InputDocumentFileLocation{ID: 1}).
			Stream(ctx, output)
		require.NoError(t, err)
		require.Zero(t, dc2.Snapshot().Decreases)
		require.Zero(t, dc4.Snapshot().Decreases)
	})
	t.Run("PerDC", func(t *testing.T) {
		output := new(bytes.Buffer)
		_, err := client.DownloadVia(floodOncePool(t, data), &tg.InputDocumentFileLocation{ID: 1}).
			WithScheduler(dc2).
			Stream(ctx, output)
		require.NoError(t, err)
		require.Equal(t, data, output.Bytes())

		// Flood wait on DC 2 shrinks its window only.
		require.EqualValues(t, 1, dc2.Snapshot().Decreases)
		require.Equal(t, 4, dc2.Snapshot().Window)
		require.Zero(t, dc2.Snapshot().InFlight)
		require.Zero(t, dc4.Snapshot().Decreases)
		require.Equal(t, 8, dc4.Snapshot().Window)
	})
}

func TestClientDownloadSchedulerHomeDC(t *testing.T) {
	ctx := context.Background()
	data := []byte("data from home DC")
	calls := 0
	client := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		_, ok := body.(*tg.UploadGetFileRequest)
		require.True(t, ok)
		calls++
		if calls == 1 {
			return nil, tgerr.New(420, "FLOOD_WAIT_0")
		}
		return &tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: data}, nil
	})
	client.downloadSchedulers = newDownloadSchedulers(&downloader.SchedulerOptions{Min: 1, Start: 8, Max: 16})
	const home, other = 2, 4
	client.session = pool.NewSyncSession(pool.Session{DC: home})

	output := new(bytes.Buffer)
	_, err := client.Download(&tg.InputDocumentFileLocation{ID: 1}).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	require.EqualValues(t, 1, client.DownloadScheduler(home).Snapshot().Decreases,
		"Download uses Scheduler of home DC")
	require.Zero(t, client.DownloadScheduler(other).Snapshot().Decreases)
}

func TestClientDownloadSchedulerDisabled(t *testing.T) {
	client := NewClient(TestAppID, TestAppHash, Options{})
	require.Nil(t, client.DownloadScheduler(2))
}
