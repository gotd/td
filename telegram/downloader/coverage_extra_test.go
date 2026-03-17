package downloader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestRetryLimitErrWraps(t *testing.T) {
	baseErr := errors.New("boom")
	err := retryLimitErr("download", 3, baseErr)
	require.ErrorIs(t, err, baseErr)
	require.Contains(t, err.Error(), "download: retry limit reached (3)")
}

func TestRetryRequestTimeoutLimit(t *testing.T) {
	var attempts []int
	calls := 0

	_, err := retryRequest[int](
		context.Background(),
		"read chunk",
		func(attempt int, err error) {
			require.Error(t, err)
			attempts = append(attempts, attempt)
		},
		func() (int, error) {
			calls++
			return 0, tgerr.New(500, tg.ErrTimeout)
		},
	)
	require.Error(t, err)
	require.Equal(t, maxRetryAttempts, calls)
	require.Len(t, attempts, maxRetryAttempts-1)
	require.Equal(t, maxRetryAttempts-1, attempts[len(attempts)-1])
	require.Contains(t, err.Error(), "read chunk: retry limit reached")
}

func TestCDNPlanValidationAndSplit(t *testing.T) {
	_, err := buildCDNRequestPlan(0, 0)
	require.ErrorContains(t, err, "invalid CDN limit")

	_, err = buildCDNRequestPlan(-1, cdnMinChunk)
	require.ErrorContains(t, err, "invalid CDN offset")

	_, err = buildCDNRequestPlan(1, cdnMinChunk)
	require.ErrorContains(t, err, "must be divisible")

	_, err = buildCDNRequestPlan(0, cdnMinChunk+1)
	require.ErrorContains(t, err, "must be divisible")

	plan, err := buildCDNRequestPlan(cdnMaxChunk-int64(cdnMinChunk), 2*cdnMinChunk)
	require.NoError(t, err)
	require.Equal(t, []cdnRequestRange{
		{offset: cdnMaxChunk - int64(cdnMinChunk), limit: cdnMinChunk},
		{offset: cdnMaxChunk, limit: cdnMinChunk},
	}, plan)

	require.Zero(t, largestCDNValidLimit(cdnMinChunk-1))
	require.Equal(t, cdnMinChunk, largestCDNValidLimit(cdnMinChunk))
}

func TestWebReportRetryAndHashes(t *testing.T) {
	var events []RetryEvent
	w := web{
		retryHandler: func(event RetryEvent) {
			events = append(events, event)
		},
	}

	w.reportRetry("noop", 0, errors.New("skip"))
	w.reportRetry("noop", 1, nil)
	require.Empty(t, events)

	retryErr := errors.New("retry")
	w.reportRetry("getWebFile", 2, retryErr)
	require.Len(t, events, 1)
	require.Equal(t, "getWebFile", events[0].Operation)
	require.Equal(t, 2, events[0].Attempt)
	require.ErrorIs(t, events[0].Err, retryErr)

	hashes, err := w.Hashes(context.Background(), 0)
	require.Nil(t, hashes)
	require.ErrorIs(t, err, errHashesNotSupported)
}

func TestBuilderToPath(t *testing.T) {
	ctx := context.Background()
	data := []byte("hello downloader")
	m := &mock{data: data}

	path := filepath.Join(t.TempDir(), "out.bin")
	typ, err := NewDownloader().
		Download(m, nil).
		WithThreads(2).
		ToPath(ctx, path)
	require.NoError(t, err)
	require.Nil(t, typ)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, content)
}

func TestBuilderToPathCreateError(t *testing.T) {
	ctx := context.Background()
	m := &mock{data: []byte("x")}

	path := filepath.Join(t.TempDir(), "missing", "out.bin")
	_, err := NewDownloader().Download(m, nil).ToPath(ctx, path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create output file")
}

func TestDownloaderWithRetryHandlerAndRedirectError(t *testing.T) {
	var called bool
	handler := func(RetryEvent) {
		called = true
	}

	d := NewDownloader()
	require.Same(t, d, d.WithRetryHandler(handler))
	require.NotNil(t, d.retryHandler)
	d.retryHandler(RetryEvent{Operation: "x", Attempt: 1, Err: errors.New("e")})
	require.True(t, called)

	err := (&RedirectError{Redirect: &tg.UploadFileCDNRedirect{DCID: 203}}).Error()
	require.Equal(t, "redirect to CDN DC 203", err)
}
