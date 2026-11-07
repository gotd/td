package telegram

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

// dataPool returns pool that answers one upload.getFile with data.
func dataPool(t *testing.T, data []byte) *tgmock.Mock {
	t.Helper()
	pool := tgmock.NewRequire(t)
	pool.ExpectFunc(func(b bin.Encoder) {
		_, ok := b.(*tg.UploadGetFileRequest)
		require.True(t, ok)
	}).ThenResult(&tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: data})
	return pool
}

func TestClientDownloaderMetrics(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	opts := downloader.SchedulerOptions{Min: 1, Start: 8, Max: 16}
	client := NewClient(TestAppID, TestAppHash, Options{
		MeterProvider:     sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
		DownloadScheduler: &opts,
	})
	m := &testMetrics{t: t, reader: reader}
	data := []byte("data")
	loc := &tg.InputDocumentFileLocation{ID: 1}

	t.Run("SharedInstruments", func(t *testing.T) {
		for range 2 {
			_, err := client.DownloadVia(dataPool(t, data), loc).Stream(ctx, new(bytes.Buffer))
			require.NoError(t, err)
		}
		require.EqualValues(t, 2, m.value("gotd.downloader.downloads", strAttr("result", "ok")),
			"downloads of the client use its MeterProvider")
	})
	t.Run("CopyPerCall", func(t *testing.T) {
		s := downloader.NewScheduler(downloader.SchedulerOptions{Start: 8})
		_ = client.Downloader().WithScheduler(s)

		// Another downloader of the client doesn't get the scheduler set above.
		_, err := client.Downloader().
			Download(tg.NewClient(floodOncePool(t, data)), loc).
			Stream(ctx, new(bytes.Buffer))
		require.NoError(t, err)
		require.Zero(t, s.Snapshot().Decreases)
	})
	t.Run("SchedulerGauges", func(t *testing.T) {
		client.DownloadScheduler(2)
		client.DownloadScheduler(4)
		require.EqualValues(t, opts.Start, m.value("gotd.downloader.scheduler.window", dcAttr(2)))
		require.EqualValues(t, opts.Start, m.value("gotd.downloader.scheduler.window", dcAttr(4)))
		require.Equal(t, 2, client.downloadSchedulers.registered())

		// Run returned: schedulers stop reporting until used again.
		client.downloadSchedulers.unregister()
		require.Zero(t, client.downloadSchedulers.registered())
		require.Zero(t, m.value("gotd.downloader.scheduler.window", dcAttr(2)))

		client.DownloadScheduler(2)
		require.Equal(t, 1, client.downloadSchedulers.registered())
		require.EqualValues(t, opts.Start, m.value("gotd.downloader.scheduler.window", dcAttr(2)))
	})
}

func TestClientDownloaderNoMetrics(t *testing.T) {
	opts := downloader.SchedulerOptions{Start: 8}
	client := NewClient(TestAppID, TestAppHash, Options{DownloadScheduler: &opts})
	require.NotNil(t, client.DownloadScheduler(2))
	require.Zero(t, client.downloadSchedulers.registered(), "no meter, no registrations")
	require.NotSame(t, client.Downloader(), client.Downloader())
	// Client not created by NewClient.
	require.NotNil(t, (&Client{}).Downloader())
}
