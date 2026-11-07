package telegram

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func TestClientDownloadVia(t *testing.T) {
	ctx := context.Background()
	// Primary connection must not be used.
	client := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		return nil, tgerr.New(400, "PRIMARY_USED")
	})

	data := []byte("data from pool")
	pool := tgmock.NewRequire(t)
	pool.ExpectFunc(func(b bin.Encoder) {
		_, ok := b.(*tg.UploadGetFileRequest)
		require.True(t, ok)
	}).ThenResult(&tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: data})

	output := new(bytes.Buffer)
	_, err := client.DownloadVia(pool, &tg.InputDocumentFileLocation{ID: 1}).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	var _ downloader.CDNProvider = client.CDNProvider()
	_, ok := downloader.Client(downloadClient{client: client, raw: tg.NewClient(pool)}).(downloader.CDNProvider)
	require.True(t, ok, "DownloadVia client must serve CDN redirects")
}
