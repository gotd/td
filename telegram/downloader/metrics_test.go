package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/gotd/log"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// testMetrics collects metrics of a test meter provider.
type testMetrics struct {
	t      *testing.T
	reader *sdkmetric.ManualReader
	mp     *sdkmetric.MeterProvider
}

func newTestMetrics(t *testing.T) *testMetrics {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	return &testMetrics{
		t:      t,
		reader: reader,
		mp:     sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
	}
}

func matches(set attribute.Set, kv []attribute.KeyValue) bool {
	for _, want := range kv {
		got, ok := set.Value(want.Key)
		if !ok || got != want.Value {
			return false
		}
	}
	return true
}

// value returns sum of data points of int64 sum or gauge metric name with
// given attributes, and whether the metric was found.
func (m *testMetrics) value(name string, kv ...attribute.KeyValue) (int64, bool) {
	m.t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(m.t, m.reader.Collect(context.Background(), &rm))
	var (
		total int64
		found bool
	)
	for _, sm := range rm.ScopeMetrics {
		for _, mt := range sm.Metrics {
			if mt.Name != name {
				continue
			}
			var points []metricdata.DataPoint[int64]
			switch data := mt.Data.(type) {
			case metricdata.Sum[int64]:
				points = data.DataPoints
			case metricdata.Gauge[int64]:
				points = data.DataPoints
			default:
				m.t.Fatalf("unexpected type %T of %s", mt.Data, name)
			}
			for _, p := range points {
				if matches(p.Attributes, kv) {
					total += p.Value
					found = true
				}
			}
		}
	}
	return total, found
}

func (m *testMetrics) count(name string, kv ...attribute.KeyValue) int64 {
	m.t.Helper()
	v, _ := m.value(name, kv...)
	return v
}

func TestMetricsCDN(t *testing.T) {
	ctx := context.Background()
	m := newTestMetrics(t)
	base := newSchedulerCDNMock(t)
	provider := &cdnFloodProvider{mock: base}
	provider.floodOnce.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithMeterProvider(m.mp).
		WithAllowCDN(true).
		Download(provider, nil).
		WithScheduler(NewScheduler(SchedulerOptions{})).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, base.data, output.Bytes())

	kind := attribute.String("kind", "cdn")
	require.EqualValues(t, 1, m.count("gotd.downloader.cdn.redirects"))
	require.EqualValues(t, 1, m.count("gotd.downloader.requests", kind, attribute.String("result", "flood_wait")))
	require.Positive(t, m.count("gotd.downloader.requests", kind, attribute.String("result", "ok")))
	require.EqualValues(t, len(base.data), m.count("gotd.downloader.bytes", kind))
	require.EqualValues(t, 1, m.count("gotd.downloader.scheduler.backoffs", attribute.String("reason", "flood_wait")))
	require.EqualValues(t, 1, m.count("gotd.downloader.downloads", attribute.String("result", "ok")))
	require.Zero(t, m.count("gotd.downloader.requests", attribute.String("kind", "master")))
}

func TestMetricsRefresh(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	data := make([]byte, part*4)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)
	newClient := func(expireAfter int) *refClient {
		mk := &mock{data: data, hashes: mockHashes{ranges: countHashes(data, part)}}
		return &refClient{noCDNClient: noCDNClient{base: mk}, expireAfter: expireAfter}
	}
	stale := &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{0xff}}
	result := func(r string) attribute.KeyValue { return attribute.String("result", r) }

	t.Run("OK", func(t *testing.T) {
		m := newTestMetrics(t)
		c := newClient(0)
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(c, stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				return c.current(), nil
			}).
			Stream(ctx, io.Discard)
		require.NoError(t, err)
		require.EqualValues(t, 1, m.count("gotd.downloader.file_reference.refreshes", result("ok")))
		require.EqualValues(t, 1, m.count("gotd.downloader.retries",
			attribute.String("operation", RetryOperationFileReference)))
	})
	t.Run("NotFound", func(t *testing.T) {
		m := newTestMetrics(t)
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(newClient(0), stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				return nil, errors.Wrap(ErrFileNotFound, "message deleted")
			}).
			Stream(ctx, io.Discard)
		require.ErrorIs(t, err, ErrFileNotFound)
		require.EqualValues(t, 1, m.count("gotd.downloader.file_reference.refreshes", result("not_found")))
		require.EqualValues(t, 1, m.count("gotd.downloader.downloads", result("error")))
	})
	t.Run("Error", func(t *testing.T) {
		m := newTestMetrics(t)
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(newClient(0), stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				return nil, errors.New("boom")
			}).
			Stream(ctx, io.Discard)
		require.Error(t, err)
		require.EqualValues(t, 1, m.count("gotd.downloader.file_reference.refreshes", result("error")))
	})
	t.Run("Limit", func(t *testing.T) {
		m := newTestMetrics(t)
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(newClient(0), stale).
			WithRefresh(func(ctx context.Context) (tg.InputFileLocationClass, error) {
				// Always stale.
				return &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{0xfe}}, nil
			}).
			Stream(ctx, io.Discard)
		require.Error(t, err)
		require.EqualValues(t, maxRefreshAttempts, m.count("gotd.downloader.file_reference.refreshes", result("ok")))
		require.EqualValues(t, 1, m.count("gotd.downloader.file_reference.refreshes", result("limit")))
	})
}

func TestMetricsResumeVerify(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	data := make([]byte, part*4)
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)
	newClient := func(hashed []byte) *noCDNClient {
		return &noCDNClient{base: &mock{data: data, hashes: mockHashes{ranges: countHashes(hashed, part)}}}
	}
	result := func(r string) attribute.KeyValue { return attribute.String("result", r) }

	t.Run("Resume", func(t *testing.T) {
		m := newTestMetrics(t)
		out := new(bytes.Buffer)
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(newClient(data), nil).
			WithOffset(part).
			Stream(ctx, out)
		require.NoError(t, err)
		require.Equal(t, data[part:], out.Bytes())
		require.EqualValues(t, 1, m.count("gotd.downloader.resumes"))
		require.EqualValues(t, len(data)-part, m.count("gotd.downloader.bytes", attribute.String("kind", "master")))
	})
	t.Run("HashMismatch", func(t *testing.T) {
		m := newTestMetrics(t)
		other := bytes.Repeat([]byte{1}, len(data))
		_, err := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp).
			Download(newClient(other), nil).
			WithVerify(true).
			Stream(ctx, io.Discard)
		require.ErrorIs(t, err, ErrHashMismatch)
		require.EqualValues(t, 1, m.count("gotd.downloader.hash.mismatches", attribute.String("kind", "verify")))
	})
	t.Run("Verify", func(t *testing.T) {
		m := newTestMetrics(t)
		d := NewDownloader().WithPartSize(part).WithMeterProvider(m.mp)
		verified, err := d.Download(newClient(data), nil).Verify(ctx, bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err)
		require.EqualValues(t, len(data), verified)

		local := append([]byte(nil), data...)
		local[part+1] ^= 0xff
		verified, err = d.Download(newClient(data), nil).Verify(ctx, bytes.NewReader(local), int64(len(local)))
		require.NoError(t, err)
		require.EqualValues(t, part, verified)

		local[0] ^= 0xff
		verified, err = d.Download(newClient(data), nil).Verify(ctx, bytes.NewReader(local), int64(len(local)))
		require.NoError(t, err)
		require.Zero(t, verified)

		require.EqualValues(t, 1, m.count("gotd.downloader.verify", result("complete")))
		require.EqualValues(t, 1, m.count("gotd.downloader.verify", result("partial")))
		require.EqualValues(t, 1, m.count("gotd.downloader.verify", result("none")))
	})
}

func TestSchedulerInstrument(t *testing.T) {
	m := newTestMetrics(t)
	s := NewScheduler(SchedulerOptions{Min: 1, Start: 8, Max: 16})
	dc := attribute.Int("dc", 2)
	reg, err := s.Instrument(m.mp.Meter("test"), dc)
	require.NoError(t, err)

	window, ok := m.value("gotd.downloader.scheduler.window", dc)
	require.True(t, ok)
	require.EqualValues(t, 8, window)
	require.EqualValues(t, 0, m.count("gotd.downloader.scheduler.in_flight", dc))
	require.EqualValues(t, 0, m.count("gotd.downloader.scheduler.decreases", dc))

	ctx := context.Background()
	epoch, err := s.acquire(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, m.count("gotd.downloader.scheduler.in_flight", dc))
	s.release(ctx, epoch, errTimeoutForTest())
	require.EqualValues(t, 4, m.count("gotd.downloader.scheduler.window", dc))
	require.EqualValues(t, 1, m.count("gotd.downloader.scheduler.decreases", dc))

	require.NoError(t, reg.Unregister())
	_, ok = m.value("gotd.downloader.scheduler.window", dc)
	require.False(t, ok, "unregistered")
}

func errTimeoutForTest() error {
	return context.DeadlineExceeded
}

// recordLogger records log messages.
type recordLogger struct {
	mux  sync.Mutex
	msgs []string
}

func (l *recordLogger) Enabled(context.Context, log.Level) bool { return true }

func (l *recordLogger) Log(_ context.Context, _ log.Level, msg string, _ ...log.Attr) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.msgs = append(l.msgs, msg)
}

func TestLoggerCDN(t *testing.T) {
	lg := &recordLogger{}
	base := newSchedulerCDNMock(t)
	_, err := NewDownloader().
		WithLogger(lg).
		WithAllowCDN(true).
		Download(&cdnFloodProvider{mock: base}, nil).
		Stream(context.Background(), io.Discard)
	require.NoError(t, err)
	require.Contains(t, lg.msgs, "Redirected to CDN")
}

func TestMetricsRefreshShared(t *testing.T) {
	// Workers that fail with the same stale location share one refresh, and
	// only the refresh that ran is counted.
	const workers = 8
	m := newTestMetrics(t)
	ctx := newObserver(m.mp, nil).with(context.Background())

	stale := &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{0}}
	fresh := &tg.InputDocumentFileLocation{ID: 1, FileReference: []byte{1}}
	var refreshes atomic.Int32
	c := master{location: stale}
	c.ref = &locationRef{location: stale, refresh: func(context.Context) (tg.InputFileLocationClass, error) {
		refreshes.Add(1)
		return fresh, nil
	}}

	var failed sync.WaitGroup
	failed.Add(workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			first := true
			_, err := withRefresh(ctx, c, func(loc tg.InputFileLocationClass) (int, error) {
				if loc == stale {
					if first {
						first = false
						// Every worker fails with the stale location
						// before any refresh.
						failed.Done()
						failed.Wait()
					}
					return 0, tgerr.New(400, "FILE_REFERENCE_EXPIRED")
				}
				return 1, nil
			})
			require.NoError(t, err)
		})
	}
	wg.Wait()
	require.EqualValues(t, 1, refreshes.Load())
	require.EqualValues(t, 1, m.count("gotd.downloader.file_reference.refreshes",
		attribute.String("result", "ok")))
}

func TestMetricsRequestDurationBuckets(t *testing.T) {
	m := newTestMetrics(t)
	o := newObserver(m.mp, nil)
	ctx := o.with(context.Background())
	o.request(ctx, "master", time.Now().Add(-30*time.Millisecond), 10, nil)

	var rm metricdata.ResourceMetrics
	require.NoError(t, m.reader.Collect(context.Background(), &rm))
	for _, sm := range rm.ScopeMetrics {
		for _, mt := range sm.Metrics {
			if mt.Name != "gotd.downloader.request.duration" {
				continue
			}
			dp := mt.Data.(metricdata.Histogram[float64]).DataPoints[0]
			require.Equal(t, []float64{
				0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60,
			}, dp.Bounds, "second-based buckets")
			return
		}
	}
	t.Fatal("no request duration recorded")
}
