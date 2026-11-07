package updates

import (
	"context"
	"time"

	"github.com/gotd/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "github.com/gotd/td/telegram/updates"

// Values of the "kind" attribute.
const (
	kindCommon  = "common"
	kindChannel = "channel"
)

// Values of the "result" attribute of difference metrics.
const (
	diffEmpty        = "empty"
	diffSlice        = "slice"
	diffFull         = "full"
	diffTooLong      = "too_long"
	diffError        = "error"
	diffInaccessible = "inaccessible"
)

// Values of the "action" attribute of gap metrics.
const (
	gapDetected        = "detected"
	gapResolvedWaiting = "resolved_by_waiting"
	gapResolvedPending = "resolved_by_pending"
)

// Values of the "action" attribute of updates received before Run.
const (
	earlyPassed    = "passed"
	earlyBuffered  = "buffered"
	earlyDropped   = "dropped"
	earlyReplayed  = "replayed"
	earlyPassedRaw = "passed_raw"
)

// Values of the "source" and "result" attributes of channel tracking.
const (
	trackSourceAPI     = "api"
	trackSourceStored  = "stored"
	trackSourceUpdate  = "update"
	trackSourceTooLong = "too_long"

	trackStarted  = "started"
	trackAlready  = "already"
	trackPending  = "pending"
	trackStopped  = "stopped"
	trackError    = "error"
	trackCanceled = "canceled"
)

// Values of the "source" attribute of dispatched updates.
const (
	sourceLive       = "live"
	sourceDifference = "difference"
)

// Values of the "op" attribute of ErrStopped returns.
const (
	opHandle         = "handle"
	opHandleAffected = "handle_affected"
	opTrackChannel   = "track_channel"
)

type gapKey struct {
	kind, seq, action string
}

// updatesMetrics are the OTel instruments of the manager.
//
// A nil *updatesMetrics records nothing, so states created without it (e.g.
// in tests) work as before.
type updatesMetrics struct {
	diffRequests  metric.Int64Counter
	diffDuration  metric.Float64Histogram
	diffTriggers  metric.Int64Counter
	gaps          metric.Int64Counter
	tooLong       metric.Int64Counter
	gapSize       metric.Int64Histogram
	gapMessages   metric.Int64Histogram
	beforeRun     metric.Int64Counter
	track         metric.Int64Counter
	tracked       metric.Int64UpDownCounter
	inaccessible  metric.Int64Counter
	stoppedCalls  metric.Int64Counter
	dispatched    metric.Int64Counter
	dispatchAttrs map[string]metric.MeasurementOption
	gapAttrs      map[gapKey]metric.MeasurementOption
}

func attrs(kv ...attribute.KeyValue) metric.MeasurementOption {
	return metric.WithAttributeSet(attribute.NewSet(kv...))
}

// newUpdatesMetrics creates instruments of mp. An instrument that can't be
// created is logged and replaced by a no-op one.
func newUpdatesMetrics(mp metric.MeterProvider, lg log.Helper) *updatesMetrics {
	meter := mp.Meter(meterName)
	fallback := noop.NewMeterProvider().Meter(meterName)
	warn := func(name string, err error) {
		lg.Warn(context.Background(), "Failed to create metric", log.String("name", name), log.Error(err))
	}
	counter := func(name, unit, desc string) metric.Int64Counter {
		opts := []metric.Int64CounterOption{metric.WithUnit(unit), metric.WithDescription(desc)}
		c, err := meter.Int64Counter(name, opts...)
		if err != nil {
			warn(name, err)
			c, _ = fallback.Int64Counter(name, opts...)
		}
		return c
	}
	intHist := func(name, unit, desc string) metric.Int64Histogram {
		opts := []metric.Int64HistogramOption{metric.WithUnit(unit), metric.WithDescription(desc)}
		h, err := meter.Int64Histogram(name, opts...)
		if err != nil {
			warn(name, err)
			h, _ = fallback.Int64Histogram(name, opts...)
		}
		return h
	}

	m := &updatesMetrics{
		diffRequests: counter("gotd.updates.difference.requests", "{request}",
			"updates.getDifference and updates.getChannelDifference calls by kind and result"),
		diffTriggers: counter("gotd.updates.difference.triggers", "{request}",
			"Reasons why the difference was requested"),
		gaps: counter("gotd.updates.gaps", "{gap}",
			"Sequence gaps detected and resolved without getDifference"),
		tooLong: counter("gotd.updates.channel.too_long", "{gap}",
			"updates.channelDifferenceTooLong received, reported via OnChannelGap"),
		gapSize: intHist("gotd.updates.channel.gap.size", "{pts}",
			"Pts skipped by updates.channelDifferenceTooLong (RemotePts - LocalPts)"),
		gapMessages: intHist("gotd.updates.channel.gap.messages", "{message}",
			"Latest messages passed in ChannelGap.Messages"),
		beforeRun: counter("gotd.updates.before_run", "{batch}",
			"Updates received before Run by action (Config.BeforeRun)"),
		track: counter("gotd.updates.channel.track", "{request}",
			"Channel tracking starts by source and result"),
		inaccessible: counter("gotd.updates.channel.inaccessible", "{channel}",
			"Channels removed from tracking after CHANNEL_PRIVATE"),
		stoppedCalls: counter("gotd.updates.stopped", "{call}",
			"Calls that returned ErrStopped because Run stopped"),
		dispatched: counter("gotd.updates.dispatched", "{update}",
			"Updates passed to Handler by source"),
	}
	{
		name := "gotd.updates.difference.duration"
		opts := []metric.Float64HistogramOption{
			metric.WithUnit("s"),
			metric.WithDescription("Duration of updates.getDifference and updates.getChannelDifference calls"),
		}
		h, err := meter.Float64Histogram(name, opts...)
		if err != nil {
			warn(name, err)
			h, _ = fallback.Float64Histogram(name, opts...)
		}
		m.diffDuration = h
	}
	{
		name := "gotd.updates.channels.tracked"
		opts := []metric.Int64UpDownCounterOption{
			metric.WithUnit("{channel}"),
			metric.WithDescription("Channels tracked by running managers"),
		}
		c, err := meter.Int64UpDownCounter(name, opts...)
		if err != nil {
			warn(name, err)
			c, _ = fallback.Int64UpDownCounter(name, opts...)
		}
		m.tracked = c
	}

	// Attributes of frequent measurements are built once.
	m.dispatchAttrs = map[string]metric.MeasurementOption{}
	for _, source := range []string{sourceLive, sourceDifference} {
		m.dispatchAttrs[source] = attrs(attribute.String("source", source))
	}
	m.gapAttrs = map[gapKey]metric.MeasurementOption{}
	for _, kind := range []string{kindCommon, kindChannel} {
		for _, seq := range []string{"pts", "qts", "seq"} {
			for _, action := range []string{gapDetected, gapResolvedWaiting, gapResolvedPending} {
				m.gapAttrs[gapKey{kind, seq, action}] = attrs(
					attribute.String("kind", kind),
					attribute.String("seq", seq),
					attribute.String("action", action),
				)
			}
		}
	}
	return m
}

// difference records a finished difference request.
func (m *updatesMetrics) difference(ctx context.Context, kind, result string, start time.Time) {
	if m == nil {
		return
	}
	opt := attrs(attribute.String("kind", kind), attribute.String("result", result))
	m.diffRequests.Add(ctx, 1, opt)
	m.diffDuration.Record(ctx, time.Since(start).Seconds(), opt)
}

// differenceTrigger records why the difference was requested. Reasons are
// constants of this package, so cardinality is bounded.
func (m *updatesMetrics) differenceTrigger(ctx context.Context, kind, reason string) {
	if m == nil {
		return
	}
	m.diffTriggers.Add(ctx, 1, attrs(attribute.String("kind", kind), attribute.String("reason", reason)))
}

func (m *updatesMetrics) gap(ctx context.Context, kind, seq, action string) {
	if m == nil {
		return
	}
	opt, ok := m.gapAttrs[gapKey{kind, seq, action}]
	if !ok {
		opt = attrs(attribute.String("kind", kind), attribute.String("seq", seq), attribute.String("action", action))
	}
	m.gaps.Add(ctx, 1, opt)
}

// channelTooLong records updates.channelDifferenceTooLong. dialog is "ok",
// "stale" (pts did not advance) or "invalid" (no dialog pts).
func (m *updatesMetrics) channelTooLong(ctx context.Context, dialog string, gap ChannelGap) {
	if m == nil {
		return
	}
	m.tooLong.Add(ctx, 1, attrs(attribute.String("dialog", dialog)))
	if gap.RemotePts > 0 {
		m.gapSize.Record(ctx, int64(gap.RemotePts-gap.LocalPts))
	}
	m.gapMessages.Record(ctx, int64(len(gap.Messages)))
}

func (m *updatesMetrics) beforeRunAction(ctx context.Context, action string, n int) {
	if m == nil || n == 0 {
		return
	}
	m.beforeRun.Add(ctx, int64(n), attrs(attribute.String("action", action)))
}

func (m *updatesMetrics) trackChannel(ctx context.Context, source, result string) {
	if m == nil {
		return
	}
	m.track.Add(ctx, 1, attrs(attribute.String("source", source), attribute.String("result", result)))
}

func (m *updatesMetrics) trackedDelta(ctx context.Context, n int) {
	if m == nil || n == 0 {
		return
	}
	m.tracked.Add(ctx, int64(n))
}

func (m *updatesMetrics) channelInaccessible(ctx context.Context) {
	if m == nil {
		return
	}
	m.inaccessible.Add(ctx, 1)
}

func (m *updatesMetrics) stopped(ctx context.Context, op string) {
	if m == nil {
		return
	}
	m.stoppedCalls.Add(ctx, 1, attrs(attribute.String("op", op)))
}

func (m *updatesMetrics) dispatch(ctx context.Context, source string, n int) {
	if m == nil || n == 0 {
		return
	}
	m.dispatched.Add(ctx, int64(n), m.dispatchAttrs[source])
}

// diffSourceKey marks contexts of updates that came from a difference.
type diffSourceKey struct{}

func withDifferenceSource(ctx context.Context) context.Context {
	return context.WithValue(ctx, diffSourceKey{}, true)
}

// updateSource returns the source of updates handled with ctx.
func updateSource(ctx context.Context) string {
	if v, _ := ctx.Value(diffSourceKey{}).(bool); v {
		return sourceDifference
	}
	return sourceLive
}
