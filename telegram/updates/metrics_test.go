package updates

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"

	"github.com/gotd/td/tg"
)

// testMeter collects metrics of a manager.
type testMeter struct {
	t      *testing.T
	reader *sdkmetric.ManualReader
	mp     *sdkmetric.MeterProvider
}

func newTestMeter(t *testing.T) *testMeter {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	return &testMeter{
		t:      t,
		reader: reader,
		mp:     sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
	}
}

func metricKey(name string, kv []attribute.KeyValue) string {
	parts := make([]string, 0, len(kv))
	for _, a := range kv {
		parts = append(parts, string(a.Key)+"="+a.Value.Emit())
	}
	sort.Strings(parts)
	return name + "{" + strings.Join(parts, ",") + "}"
}

// values returns sums of counters, and counts and sums of histograms, keyed by
// `name{k=v,...}` with sorted attributes.
func (m *testMeter) values() map[string]int64 {
	m.t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(m.t, m.reader.Collect(context.Background(), &rm))
	out := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, mt := range sm.Metrics {
			switch data := mt.Data.(type) {
			case metricdata.Sum[int64]:
				for _, dp := range data.DataPoints {
					out[metricKey(mt.Name, dp.Attributes.ToSlice())] += dp.Value
				}
			case metricdata.Histogram[int64]:
				for _, dp := range data.DataPoints {
					k := metricKey(mt.Name, dp.Attributes.ToSlice())
					out[k+".count"] += int64(dp.Count)
					out[k+".sum"] += dp.Sum
				}
			case metricdata.Histogram[float64]:
				for _, dp := range data.DataPoints {
					out[metricKey(mt.Name, dp.Attributes.ToSlice())+".count"] += int64(dp.Count)
				}
			}
		}
	}
	return out
}

func (m *testMeter) eventually(key string, want int64) {
	m.t.Helper()
	require.Eventually(m.t, func() bool {
		return m.values()[key] == want
	}, 5*time.Second, time.Millisecond, "%s: want %d", key, want)
}

func (m *testMeter) updatesMetrics() *updatesMetrics {
	return newUpdatesMetrics(m.mp, log.For(logzap.New(zaptest.NewLogger(m.t))))
}

func TestMetricsBeforeRun(t *testing.T) {
	ctx := context.Background()
	newManager := func(t *testing.T, mode BeforeRunMode, size int) (*Manager, *testMeter) {
		t.Helper()
		tm := newTestMeter(t)
		storage := newMemStorage()
		require.NoError(t, storage.SetState(ctx, trackUserID, State{Pts: 1, Qts: 1, Seq: 1, Date: 1}))
		return New(Config{
			Handler:             &beforeRunHandler{},
			Storage:             storage,
			Logger:              logzap.New(zaptest.NewLogger(t)),
			MeterProvider:       tm.mp,
			BeforeRun:           mode,
			BeforeRunBufferSize: size,
		}), tm
	}
	run := func(t *testing.T, m *Manager) {
		t.Helper()
		ctx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() { done <- m.Run(ctx, &trackAPI{}, trackUserID, AuthOptions{}) }()
		t.Cleanup(func() {
			cancel()
			<-done
		})
	}

	t.Run("Pass", func(t *testing.T) {
		m, tm := newManager(t, BeforeRunPass, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		tm.eventually("gotd.updates.before_run{action=passed}", 1)
	})
	t.Run("Buffer", func(t *testing.T) {
		m, tm := newManager(t, BeforeRunBuffer, 1)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		require.NoError(t, m.Handle(ctx, newMessageUpdates(2, 3)))
		tm.eventually("gotd.updates.before_run{action=buffered}", 1)
		tm.eventually("gotd.updates.before_run{action=dropped}", 1)

		run(t, m)
		tm.eventually("gotd.updates.before_run{action=replayed}", 1)
		tm.eventually("gotd.updates.dispatched{source=live}", 1)
	})
	t.Run("Drop", func(t *testing.T) {
		m, tm := newManager(t, BeforeRunDrop, 0)
		require.NoError(t, m.Handle(ctx, newMessageUpdates(1, 2)))
		tm.eventually("gotd.updates.before_run{action=dropped}", 1)
	})
}

func TestMetricsTrackChannel(t *testing.T) {
	ctx := context.Background()
	tm := newTestMeter(t)
	tt := newTrackTest(t)
	tt.m = newTestManager(t, Config{
		Storage:       tt.storage,
		AccessHasher:  tt.hasher,
		MeterProvider: tm.mp,
	})

	require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 100))
	tm.eventually("gotd.updates.channel.track{result=pending,source=api}", 1)

	runCtx, cancel := context.WithCancel(ctx)
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- tt.m.Run(runCtx, tt.api, trackUserID, AuthOptions{
			OnStart: func(context.Context) { close(started) },
		})
	}()
	<-started
	tt.waitPts(t, 42, 100)
	tm.eventually("gotd.updates.channel.track{result=started,source=api}", 1)
	tm.eventually("gotd.updates.channels.tracked{}", 1)
	tm.eventually("gotd.updates.difference.requests{kind=common,result=empty}", 1)
	tm.eventually("gotd.updates.difference.requests{kind=channel,result=empty}", 1)
	tm.eventually("gotd.updates.difference.triggers{kind=common,reason=startup}", 1)
	tm.eventually("gotd.updates.difference.triggers{kind=channel,reason=channel-subscribe}", 1)
	tm.eventually("gotd.updates.difference.duration{kind=channel,result=empty}.count", 1)

	// Already tracked.
	require.NoError(t, tt.m.TrackChannel(ctx, 42, 4242, 100))
	tm.eventually("gotd.updates.channel.track{result=already,source=api}", 1)
	// No stored pts and no pts.
	require.Error(t, tt.m.TrackChannel(ctx, 43, 4343, 0))
	tm.eventually("gotd.updates.channel.track{result=error,source=api}", 1)

	cancel()
	<-done
	tm.eventually("gotd.updates.channels.tracked{}", 0)

	require.ErrorIs(t, tt.m.TrackChannel(ctx, 44, 4444, 100), ErrStopped)
	// Updates fill the queue first.
	var err error
	for i := 0; i < 100 && err == nil; i++ {
		err = tt.m.Handle(ctx, &tg.Updates{Date: 1})
	}
	require.ErrorIs(t, err, ErrStopped)
	tm.eventually("gotd.updates.channel.track{result=stopped,source=api}", 1)
	tm.eventually("gotd.updates.stopped{op=track_channel}", 1)
	tm.eventually("gotd.updates.stopped{op=handle}", 1)
}

func TestMetricsChannelTooLong(t *testing.T) {
	ctx := context.Background()
	tm := newTestMeter(t)
	api := &tooLongAPI{final: true}
	var r tooLongResult
	s := newTooLongChannelState(t, api, &r)
	s.metrics = tm.updatesMetrics()

	require.NoError(t, s.getDifference(ctx, "test"))
	require.Len(t, r.gaps, 1)

	tm.eventually("gotd.updates.channel.too_long{dialog=ok}", 1)
	// LocalPts 100, RemotePts 5000.
	tm.eventually("gotd.updates.channel.gap.size{}.sum", 4900)
	tm.eventually("gotd.updates.channel.gap.messages{}.sum", 3)
	tm.eventually("gotd.updates.difference.requests{kind=channel,result=too_long}", 1)
	tm.eventually("gotd.updates.difference.triggers{kind=channel,reason=test}", 1)

	t.Run("Stale", func(t *testing.T) {
		tm := newTestMeter(t)
		api := &tooLongAPI{final: true, remotePts: 50}
		var r tooLongResult
		s := newTooLongChannelState(t, api, &r)
		s.metrics = tm.updatesMetrics()

		require.NoError(t, s.getDifference(ctx, "test"))
		tm.eventually("gotd.updates.channel.too_long{dialog=stale}", 1)
		tm.eventually("gotd.updates.channel.gap.size{}.sum", 1)
	})
}

func TestMetricsGaps(t *testing.T) {
	ctx := context.Background()
	tm := newTestMeter(t)
	var applied []int
	box := newSequenceBox(sequenceConfig{
		InitialState: 1,
		Apply: func(_ context.Context, state int, _ []update) error {
			applied = append(applied, state)
			return nil
		},
		Metrics: tm.updatesMetrics(),
		Kind:    kindCommon,
		Name:    "pts",
	})

	// Gap: 2 is missing.
	require.NoError(t, box.Handle(ctx, update{State: 3, Count: 1}))
	tm.eventually("gotd.updates.gaps{action=detected,kind=common,seq=pts}", 1)
	require.NoError(t, box.Handle(ctx, update{State: 2, Count: 1}))
	tm.eventually("gotd.updates.gaps{action=resolved_by_waiting,kind=common,seq=pts}", 1)
	require.Equal(t, []int{3}, applied)
}

func TestMetricsNil(t *testing.T) {
	// States created without metrics must not panic.
	var m *updatesMetrics
	ctx := context.Background()
	m.difference(ctx, kindCommon, diffEmpty, time.Now())
	m.differenceTrigger(ctx, kindCommon, "test")
	m.gap(ctx, kindCommon, "pts", gapDetected)
	m.channelTooLong(ctx, "ok", ChannelGap{})
	m.beforeRunAction(ctx, earlyPassed, 1)
	m.trackChannel(ctx, trackSourceAPI, trackStarted)
	m.trackedDelta(ctx, 1)
	m.channelInaccessible(ctx)
	m.stopped(ctx, opHandle)
	m.dispatch(ctx, sourceLive, 1)
}

func TestMetricsDifferenceDurationBuckets(t *testing.T) {
	m := newTestMeter(t)
	um := m.updatesMetrics()
	um.difference(context.Background(), kindCommon, "empty", time.Now().Add(-30*time.Millisecond))

	var rm metricdata.ResourceMetrics
	require.NoError(t, m.reader.Collect(context.Background(), &rm))
	for _, sm := range rm.ScopeMetrics {
		for _, mt := range sm.Metrics {
			if mt.Name != "gotd.updates.difference.duration" {
				continue
			}
			dp := mt.Data.(metricdata.Histogram[float64]).DataPoints[0]
			require.Equal(t, []float64{
				0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60,
			}, dp.Bounds, "second-based buckets")
			// 30ms lands in the (0.025, 0.05] bucket.
			require.Equal(t, uint64(1), dp.BucketCounts[3])
			return
		}
	}
	t.Fatal("no difference duration recorded")
}
