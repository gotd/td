package downloader

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/syncio"
	"github.com/gotd/td/tg"
)

func TestBuilderWithSizeFileType(t *testing.T) {
	ctx := context.Background()
	// More threads than parts: workers past the known end of file must not
	// reset the file type reported by the server.
	data := resumeData(t, resumePart*2+10)
	for i := 0; i < 20; i++ {
		m := &mock{data: data, fileType: &tg.StorageFileMp4{}}
		out := new(syncio.BufWriterAt)
		typ, err := NewDownloader().WithPartSize(resumePart).
			Download(m, nil).
			WithThreads(8).
			WithSize(int64(len(data))).
			Parallel(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data, out.Bytes())
		require.Equal(t, &tg.StorageFileMp4{}, typ)
	}
}

func TestBuilderWithOffsetVerifyWindow(t *testing.T) {
	ctx := context.Background()
	// Hash windows are larger than the 4KB-aligned offset, as on real servers
	// (128KB): the first window starts before the offset.
	const window = 128 * 1024
	data := resumeData(t, window*2+100)

	for _, offset := range []int64{4096, window - 4096, window + 8192} {
		m := &mock{data: data, hashes: mockHashes{ranges: countHashes(data, window)}}
		out := new(bytes.Buffer)
		var last int64
		_, err := NewDownloader().WithPartSize(window).
			Download(m, nil).
			WithVerify(true).
			WithSize(int64(len(data))).
			WithOffset(offset).
			WithProgress(func(done, total int64) {
				require.LessOrEqual(t, done, total)
				last = done
			}).
			Stream(ctx, out)
		require.NoError(t, err, "offset %d", offset)
		require.Equal(t, data[offset:], out.Bytes(), "offset %d", offset)
		require.Equal(t, int64(len(data)), last, "offset %d", offset)
	}
}

func TestBuilderToPathTruncatesToSize(t *testing.T) {
	ctx := context.Background()
	data := resumeData(t, resumePart*2+10)
	path := filepath.Join(t.TempDir(), "file.part")
	// Stale longer partial file.
	stale := append(append([]byte{}, data[:resumePart]...), resumeData(t, resumePart*4)...)
	require.NoError(t, os.WriteFile(path, stale, 0o600))

	_, err := NewDownloader().WithPartSize(resumePart).
		Download(&mock{data: data}, nil).
		WithSize(int64(len(data))).
		WithOffset(resumePart).
		ToPath(ctx, path)
	require.NoError(t, err)
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, got)
}
