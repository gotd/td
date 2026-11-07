package downloader

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/syncio"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestSchedulerWindow(t *testing.T) {
	ctx := context.Background()
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 2, Max: 4})
	require.Equal(t, 2, s.Snapshot().Window)

	do := func(err error) {
		epoch, acquireErr := s.acquire(ctx)
		require.NoError(t, acquireErr)
		s.release(ctx, epoch, err)
	}
	ok := func(n int) {
		for range n {
			do(nil)
		}
	}
	ok(2)
	require.Equal(t, 3, s.Snapshot().Window, "grows after window of successes")
	ok(3)
	require.Equal(t, 4, s.Snapshot().Window)
	ok(100)
	require.Equal(t, 4, s.Snapshot().Window, "capped by Max")

	do(tgerr.New(420, "FLOOD_WAIT_1"))
	require.Equal(t, 2, s.Snapshot().Window, "halved on flood wait")
	do(tgerr.New(500, tg.ErrTimeout))
	require.Equal(t, 1, s.Snapshot().Window, "halved on timeout")
	do(tgerr.New(420, "FLOOD_PREMIUM_WAIT_1"))
	require.Equal(t, 1, s.Snapshot().Window, "capped by Min")
	require.EqualValues(t, 3, s.Snapshot().Decreases)

	do(tgerr.New(400, "FILE_ID_INVALID"))
	require.Equal(t, 1, s.Snapshot().Window, "other errors are neutral")
	require.Zero(t, s.Snapshot().InFlight)
}

func TestSchedulerBurstHalvesOnce(t *testing.T) {
	ctx := context.Background()
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 16, Max: 16})

	// All requests of the window are in flight and fail together.
	epochs := make([]uint64, 16)
	for i := range epochs {
		epoch, err := s.acquire(ctx)
		require.NoError(t, err)
		epochs[i] = epoch
	}
	for _, epoch := range epochs {
		s.release(ctx, epoch, tgerr.New(420, "FLOOD_WAIT_1"))
	}
	require.Equal(t, 8, s.Snapshot().Window, "halved once per burst")
	require.EqualValues(t, 1, s.Snapshot().Decreases)

	// Request sent after the decrease halves again.
	epoch, err := s.acquire(ctx)
	require.NoError(t, err)
	s.release(ctx, epoch, tgerr.New(420, "FLOOD_WAIT_1"))
	require.Equal(t, 4, s.Snapshot().Window)
}

func TestSchedulerAcquireContext(t *testing.T) {
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 1, Max: 1})
	epoch, err := s.acquire(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = s.acquire(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Zero(t, s.Snapshot().Waiting)

	s.release(context.Background(), epoch, nil)
	_, err = s.acquire(context.Background())
	require.NoError(t, err)
}

// peakClient records peak of concurrent UploadGetFile calls.
type peakClient struct {
	noCDNClient
	mux       sync.Mutex
	cur, peak int
}

func (c *peakClient) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	c.mux.Lock()
	c.cur++
	if c.cur > c.peak {
		c.peak = c.cur
	}
	c.mux.Unlock()
	time.Sleep(time.Millisecond)
	defer func() {
		c.mux.Lock()
		c.cur--
		c.mux.Unlock()
	}()
	return c.noCDNClient.UploadGetFile(ctx, request)
}

func TestSchedulerSharedByDownloads(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	data := make([]byte, part*32)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)

	c := &peakClient{noCDNClient: noCDNClient{base: &mock{data: data}}}
	s := NewScheduler(SchedulerOptions{Min: 2, Start: 2, Max: 2})
	d := NewDownloader().WithPartSize(part).WithScheduler(s)

	var g errgroup.Group
	for range 3 {
		g.Go(func() error {
			out := new(syncio.BufWriterAt)
			if _, err := d.Download(c, nil).WithThreads(8).Parallel(ctx, out); err != nil {
				return err
			}
			require.Equal(t, data, out.Bytes())
			return nil
		})
	}
	require.NoError(t, g.Wait())
	require.LessOrEqual(t, c.peak, 2, "in-flight requests are limited by shared window")
	require.Zero(t, s.Snapshot().InFlight)
}
