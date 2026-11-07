package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/syncio"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// refClient rejects requests with stale file reference, and makes current
// reference expire after expireAfter successful chunks.
type refClient struct {
	noCDNClient

	mux         sync.Mutex
	valid       byte
	served      int
	expireAfter int
}

func (c *refClient) check(loc tg.InputFileLocationClass) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	l := loc.(*tg.InputDocumentFileLocation)
	if !bytes.Equal(l.FileReference, []byte{c.valid}) {
		return tgerr.New(400, "FILE_REFERENCE_EXPIRED")
	}
	return nil
}

func (c *refClient) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	if err := c.check(request.Location); err != nil {
		return nil, err
	}
	c.mux.Lock()
	c.served++
	if c.expireAfter > 0 && c.served%c.expireAfter == 0 {
		c.valid++
	}
	c.mux.Unlock()
	return c.noCDNClient.UploadGetFile(ctx, request)
}

func (c *refClient) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	if err := c.check(request.Location); err != nil {
		return nil, err
	}
	return c.noCDNClient.UploadGetFileHashes(ctx, request)
}

func (c *refClient) current() tg.InputFileLocationClass {
	c.mux.Lock()
	defer c.mux.Unlock()
	return &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{c.valid}}
}

func TestBuilderWithRefresh(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	data := make([]byte, part*10+10)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)

	newClient := func(expireAfter int) *refClient {
		m := &mock{data: data, hashes: mockHashes{ranges: countHashes(data, part)}}
		return &refClient{noCDNClient: noCDNClient{base: m}, expireAfter: expireAfter}
	}
	stale := &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{0xff}}

	t.Run("Parallel", func(t *testing.T) {
		c := newClient(3)
		var refreshes, retries atomic.Int32
		out := new(syncio.BufWriterAt)
		_, err := NewDownloader().WithPartSize(part).
			WithRetryHandler(func(e RetryEvent) {
				if e.Operation == RetryOperationFileReference {
					retries.Add(1)
				}
			}).
			Download(c, stale).
			WithThreads(4).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				refreshes.Add(1)
				return c.current(), nil
			}).
			Parallel(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data, out.Bytes())
		require.Positive(t, refreshes.Load())
		require.Positive(t, retries.Load())
		// Initial stale location and expiry every 3 chunks of 11.
		require.LessOrEqual(t, refreshes.Load(), int32(1+11/3+1), "workers share refresh")
	})
	t.Run("Verify", func(t *testing.T) {
		c := newClient(4)
		out := new(bytes.Buffer)
		_, err := NewDownloader().WithPartSize(part).
			Download(c, stale).
			WithVerify(true).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				return c.current(), nil
			}).
			Stream(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data, out.Bytes())
	})
	t.Run("NoRefresh", func(t *testing.T) {
		_, err := NewDownloader().WithPartSize(part).
			Download(newClient(0), stale).
			Stream(ctx, io.Discard)
		require.Error(t, err)
		require.True(t, IsFileReferenceError(err))
	})
	t.Run("RefreshError", func(t *testing.T) {
		testErr := errors.New("message deleted")
		_, err := NewDownloader().WithPartSize(part).
			Download(newClient(0), stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				return nil, testErr
			}).
			Stream(ctx, io.Discard)
		require.ErrorIs(t, err, testErr)
	})
	t.Run("StillStale", func(t *testing.T) {
		var refreshes int
		_, err := NewDownloader().WithPartSize(part).
			Download(newClient(0), stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				refreshes++
				return &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{0xfe}}, nil
			}).
			Stream(ctx, io.Discard)
		require.True(t, IsFileReferenceError(err))
		require.Equal(t, maxRefreshAttempts, refreshes)
	})
}

func TestIsFileReferenceError(t *testing.T) {
	for _, typ := range []string{"FILE_REFERENCE_EXPIRED", "FILE_REFERENCE_INVALID", "FILE_REFERENCE_0_EXPIRED"} {
		require.True(t, IsFileReferenceError(errors.Wrap(tgerr.New(400, typ), "wrap")), typ)
	}
	require.False(t, IsFileReferenceError(tgerr.New(400, "FILE_ID_INVALID")))
	require.False(t, IsFileReferenceError(errors.New("FILE_REFERENCE_EXPIRED")))
}
