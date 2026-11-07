package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/syncio"
)

const resumePart = 4096

func resumeData(t *testing.T, size int) []byte {
	t.Helper()
	data := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)
	return data
}

func TestBuilderWithOffset(t *testing.T) {
	ctx := context.Background()
	data := resumeData(t, resumePart*3+100)
	offset := int64(resumePart * 2)

	t.Run("Parallel", func(t *testing.T) {
		m := &mock{data: data}
		out := new(syncio.BufWriterAt)
		_, err := out.WriteAt(data[:offset], 0)
		require.NoError(t, err)

		_, err = NewDownloader().WithPartSize(resumePart).
			Download(m, nil).
			WithThreads(4).
			WithSize(int64(len(data))).
			WithOffset(offset).
			Parallel(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data, out.Bytes())
		require.EqualValues(t, 2, m.getFileCalls.Load(), "only parts after offset")
	})
	t.Run("Stream", func(t *testing.T) {
		m := &mock{data: data}
		out := new(bytes.Buffer)
		_, err := NewDownloader().WithPartSize(resumePart).
			Download(m, nil).
			WithOffset(offset).
			Stream(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data[offset:], out.Bytes())
	})
	t.Run("Verify", func(t *testing.T) {
		m := &mock{data: data, hashes: mockHashes{ranges: countHashes(data, resumePart)}}
		out := new(bytes.Buffer)
		_, err := NewDownloader().WithPartSize(resumePart).
			Download(m, nil).
			WithVerify(true).
			WithOffset(offset).
			Stream(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data[offset:], out.Bytes())
	})
	t.Run("ToPath", func(t *testing.T) {
		m := &mock{data: data}
		path := filepath.Join(t.TempDir(), "file.part")
		require.NoError(t, os.WriteFile(path, data[:offset], 0o600))

		_, err := NewDownloader().WithPartSize(resumePart).
			Download(m, nil).
			WithOffset(offset).
			ToPath(ctx, path)
		require.NoError(t, err)
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, data, got)
	})
	t.Run("Misaligned", func(t *testing.T) {
		for _, off := range []int64{-resumePart, 100, resumePart / 2} {
			_, err := NewDownloader().WithPartSize(resumePart*2).
				Download(&mock{data: data}, nil).
				WithOffset(off).
				Stream(ctx, io.Discard)
			require.Error(t, err, "offset %d", off)
		}
	})
}

func TestBuilderWithSize(t *testing.T) {
	ctx := context.Background()
	// Exact multiple of part size: without size, workers probe past EOF.
	data := resumeData(t, resumePart*4)

	m := &mock{data: data}
	out := new(syncio.BufWriterAt)

	var (
		mux      sync.Mutex
		progress [][2]int64
	)
	_, err := NewDownloader().WithPartSize(resumePart).
		Download(m, nil).
		WithThreads(4).
		WithSize(int64(len(data))).
		WithProgress(func(done, total int64) {
			mux.Lock()
			progress = append(progress, [2]int64{done, total})
			mux.Unlock()
		}).
		Parallel(ctx, out)
	require.NoError(t, err)
	require.Equal(t, data, out.Bytes())
	require.EqualValues(t, 4, m.getFileCalls.Load(), "no requests past known size")

	require.Len(t, progress, 4)
	for i, p := range progress {
		require.Equal(t, int64(resumePart*(i+1)), p[0])
		require.Equal(t, int64(len(data)), p[1])
	}
}

func TestBuilderProgressWithOffset(t *testing.T) {
	ctx := context.Background()
	data := resumeData(t, resumePart*2+10)

	var last int64
	_, err := NewDownloader().WithPartSize(resumePart).
		Download(&mock{data: data}, nil).
		WithOffset(resumePart).
		WithProgress(func(done, total int64) {
			require.Zero(t, total)
			last = done
		}).
		Stream(ctx, io.Discard)
	require.NoError(t, err)
	require.Equal(t, int64(len(data)), last, "progress counts from file start")
}
