package downloader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestNeedPrecise(t *testing.T) {
	for _, tt := range []struct {
		offset int64
		limit  int
		want   bool
	}{
		{0, 512 * 1024, false},
		{512 * 1024, 512 * 1024, false},
		{3 * 1024 * 1024, 1024 * 1024, false},
		{8 * 1024 * 1024, 128 * 1024, false},
		{4096, 4096, false},
		{100, 512 * 1024, true},            // unaligned offset (resume)
		{0, 1000, true},                    // limit not divisible by 4 KiB
		{0, 3 * 4096, true},                // 1 MiB not divisible by limit
		{768 * 1024, 512 * 1024, true},     // crosses a 1 MiB boundary
		{1024*1024 - 4096, 8 * 1024, true}, // crosses a 1 MiB boundary
		{0, 0, true},
	} {
		require.Equal(t, tt.want, needPrecise(tt.offset, tt.limit), "offset %d limit %d", tt.offset, tt.limit)
	}
}

type flagClient struct {
	noCDNClient
	reqs []*tg.UploadGetFileRequest
}

func (c *flagClient) UploadGetFile(_ context.Context, req *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	c.reqs = append(c.reqs, req)
	return &tg.UploadFile{Bytes: make([]byte, req.Limit)}, nil
}

// Aligned parts are requested without precise, like Telegram Desktop; an
// unaligned one (resume offset) with it.
func TestMasterChunkPreciseOnlyWhenNeeded(t *testing.T) {
	c := &flagClient{}
	m := master{client: c, precise: true, location: &tg.InputDocumentFileLocation{ID: 1}}
	ctx := context.Background()

	_, err := m.Chunk(ctx, 512*1024, 512*1024)
	require.NoError(t, err)
	_, err = m.Chunk(ctx, 100, 512*1024)
	require.NoError(t, err)

	require.Len(t, c.reqs, 2)
	require.False(t, c.reqs[0].Precise)
	require.True(t, c.reqs[1].Precise)
	require.False(t, c.reqs[0].CDNSupported)

	// CDN support is sent when the download allows it.
	m.allowCDN = true
	_, err = m.Chunk(ctx, 0, 512*1024)
	require.NoError(t, err)
	require.True(t, c.reqs[2].CDNSupported)
}
