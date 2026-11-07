package downloader

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/rand"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestSchedulerBackoff(t *testing.T) {
	ctx := context.Background()
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})

	epoch, err := s.acquire(ctx)
	require.NoError(t, err)

	// Errors that are not flood waits or timeouts are ignored.
	require.Equal(t, epoch, s.backoff(ctx, epoch, tgerr.New(400, "FILE_TOKEN_INVALID")))
	require.Equal(t, 8, s.Snapshot().Window)

	flood := tgerr.New(420, "FLOOD_WAIT_1")
	epoch = s.backoff(ctx, epoch, flood)
	snap := s.Snapshot()
	require.Equal(t, 4, snap.Window)
	require.Equal(t, 1, snap.InFlight, "backoff keeps the slot")

	// Another report of the same request, retried for a long time, halves
	// window again: its epoch was refreshed by the previous report.
	epoch = s.backoff(ctx, epoch, flood)
	require.Equal(t, 2, s.Snapshot().Window)

	// Request that was sent before the last decrease does not decrease again.
	require.Equal(t, epoch, s.backoff(ctx, epoch-1, flood))
	require.Equal(t, 2, s.Snapshot().Window)

	s.release(ctx, epoch, nil)
	require.Zero(t, s.Snapshot().InFlight)

	// Repeated timeouts of one stuck request halve the window once.
	s = NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})
	epoch, err = s.acquire(ctx)
	require.NoError(t, err)
	timeout := tgerr.New(500, tg.ErrTimeout)
	for i := 0; i < 3; i++ {
		epoch = s.backoff(ctx, epoch, timeout)
	}
	require.Equal(t, 4, s.Snapshot().Window)
	require.EqualValues(t, 1, s.Snapshot().Decreases)
	s.release(ctx, epoch, timeout)
	require.EqualValues(t, 1, s.Snapshot().Decreases, "release of same request does not decrease again")
}

func TestSchedulerRetryExhausted(t *testing.T) {
	ctx := context.Background()
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})
	epoch, err := s.acquire(ctx)
	require.NoError(t, err)
	rep := &backoffReporter{sch: s, epoch: epoch}

	// A request that times out until retries are exhausted, then fails.
	calls := 0
	_, err = retryRequest(withBackoffReporter(ctx, rep), "test", nil, func() (int, error) {
		calls++
		return 0, tgerr.New(500, tg.ErrTimeout)
	})
	require.Error(t, err)
	require.Equal(t, maxRetryAttempts, calls)
	rep.release(ctx, err)

	snap := s.Snapshot()
	require.EqualValues(t, 1, snap.Decreases, "one stuck request must decrease window once")
	require.Equal(t, 4, snap.Window)
	require.Zero(t, snap.InFlight)
}

// newSchedulerCDNMock returns mock that redirects whole file to CDN.
func newSchedulerCDNMock(t *testing.T) *mock {
	t.Helper()

	data := make([]byte, defaultPartSize)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	_, err = io.ReadFull(rand.Reader, key)
	require.NoError(t, err)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	return &mock{
		data:    data,
		hashes:  mockHashes{ranges: countHashes(data, 128*1024)},
		migrate: true,
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}
}

func TestSchedulerCDNRetriedBackoff(t *testing.T) {
	ctx := context.Background()
	base := newSchedulerCDNMock(t)
	// FILE_TOKEN_INVALID from CDN makes the CDN path refresh the redirect via
	// master, and that refresh times out once: the timeout is retried inside
	// the CDN path, so the reader never sees it.
	base.tokenInvalid.Store(true)
	provider := &refreshRetryProvider{mock: base, redirect: base.redirect}
	provider.refreshTimeoutOnce.Store(true)

	s := NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})
	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(provider, nil).
		WithScheduler(s).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, base.data, output.Bytes())

	snap := s.Snapshot()
	require.EqualValues(t, 1, snap.Decreases, "retried CDN timeout must decrease window")
	require.Zero(t, snap.InFlight)
}

// cdnFloodProvider returns FLOOD_WAIT_0 from upload.getCdnFile once.
type cdnFloodProvider struct {
	*mock

	floodOnce atomic.Bool
}

func (p *cdnFloodProvider) UploadGetCDNFile(
	ctx context.Context, request *tg.UploadGetCDNFileRequest,
) (tg.UploadCDNFileClass, error) {
	if p.floodOnce.CompareAndSwap(true, false) {
		return nil, tgerr.New(420, "FLOOD_WAIT_0")
	}
	return p.mock.UploadGetCDNFile(ctx, request)
}

func (p *cdnFloodProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return p, noopCloser{}, nil
}

// TestSchedulerCDNFloodWait is regression coverage: flood waits of
// upload.getCdnFile reach the reader, so they were seen by Scheduler before
// CDN-internal retries were reported too.
func TestSchedulerCDNFloodWait(t *testing.T) {
	ctx := context.Background()
	base := newSchedulerCDNMock(t)
	provider := &cdnFloodProvider{mock: base}
	provider.floodOnce.Store(true)

	s := NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})
	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(provider, nil).
		WithScheduler(s).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, base.data, output.Bytes())
	require.EqualValues(t, 1, s.Snapshot().Decreases, "CDN flood wait must decrease window")
}
