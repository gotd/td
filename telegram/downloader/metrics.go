package downloader

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/gotd/log"

	"github.com/gotd/td/tgerr"
)

// meterName is the name of the meter of downloader metrics.
const meterName = "github.com/gotd/td/telegram/downloader"

// ErrFileNotFound means that a file reference can't be refreshed because the
// object that contains the file (message, story, photo) or the file itself is
// not found anymore.
//
// RefreshFunc should return an error that wraps it in this case, see package
// fileref.
var ErrFileNotFound = errors.New("file not found")

// Request kinds.
const (
	kindMaster = "master"
	kindCDN    = "cdn"
	kindWeb    = "web"
)

var requestKinds = [...]string{kindMaster, kindCDN, kindWeb}

// Request results.
const (
	resultOK        = "ok"
	resultFloodWait = "flood_wait"
	resultTimeout   = "timeout"
	resultCanceled  = "canceled"
	resultError     = "error"
)

var requestResults = [...]string{resultOK, resultFloodWait, resultTimeout, resultCanceled, resultError}

// File reference refresh results.
const (
	refreshOK       = "ok"
	refreshNotFound = "not_found"
	refreshError    = "error"
	refreshLimit    = "limit"
)

// Builder.Verify results.
const (
	verifyComplete = "complete"
	verifyPartial  = "partial"
	verifyNone     = "none"
	verifyError    = "error"
)

// observer reports downloader metrics and debug logs.
//
// A nil observer reports nothing: downloads without metrics and logger pay
// nothing for it.
type observer struct {
	log log.Helper

	requests         metric.Int64Counter
	bytes            metric.Int64Counter
	duration         metric.Float64Histogram
	downloads        metric.Int64Counter
	resumes          metric.Int64Counter
	cdnRedirects     metric.Int64Counter
	cdnReuploads     metric.Int64Counter
	hashMismatches   metric.Int64Counter
	refreshes        metric.Int64Counter
	backoffs         metric.Int64Counter
	verify           metric.Int64Counter
	retries          metric.Int64Counter
	retriesExhausted metric.Int64Counter

	// requestAttrs[kind][result] and kindAttrs[kind] are pre-built, so
	// reporting a request does not allocate.
	requestAttrs [len(requestKinds)][len(requestResults)]metric.MeasurementOption
	kindAttrs    [len(requestKinds)]metric.MeasurementOption
}

func attrOpt(kv ...attribute.KeyValue) metric.MeasurementOption {
	return metric.WithAttributeSet(attribute.NewSet(kv...))
}

func newObserver(mp metric.MeterProvider, lg log.Logger) *observer {
	if mp == nil {
		mp = noop.NewMeterProvider()
	}
	m := mp.Meter(meterName)
	o := &observer{log: log.For(lg)}

	// Instrument creation errors must not fail downloads: fall back to noop
	// instruments.
	nm := noop.Meter{}
	counter := func(name, unit, desc string) metric.Int64Counter {
		c, err := m.Int64Counter(name, metric.WithUnit(unit), metric.WithDescription(desc))
		if err != nil {
			c, _ = nm.Int64Counter(name)
		}
		return c
	}
	o.requests = counter("gotd.downloader.requests", "{request}",
		"Part requests by kind (master, cdn, web) and result")
	o.bytes = counter("gotd.downloader.bytes", "By", "Downloaded bytes by kind")
	o.downloads = counter("gotd.downloader.downloads", "{download}", "Finished downloads by result")
	o.resumes = counter("gotd.downloader.resumes", "{download}", "Downloads resumed from an offset")
	o.cdnRedirects = counter("gotd.downloader.cdn.redirects", "{redirect}", "Redirects to CDN DC")
	o.cdnReuploads = counter("gotd.downloader.cdn.reuploads", "{request}", "CDN file reupload requests")
	o.hashMismatches = counter("gotd.downloader.hash_mismatches", "{part}",
		"Parts that failed hash verification by kind (cdn, verify)")
	o.refreshes = counter("gotd.downloader.file_reference.refreshes", "{refresh}",
		"File reference refreshes by result (ok, not_found, error, limit)")
	o.backoffs = counter("gotd.downloader.scheduler.backoffs", "{backoff}",
		"Flood waits and timeouts reported to Scheduler by reason")
	o.verify = counter("gotd.downloader.verify", "{call}",
		"Builder.Verify calls by result (complete, partial, none, error)")
	o.retries = counter("gotd.downloader.retries", "{retry}", "Retried transient errors by operation")
	o.retriesExhausted = counter("gotd.downloader.retries_exhausted", "{request}",
		"Requests that failed after the retry limit by operation")

	h, err := m.Float64Histogram("gotd.downloader.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Part request duration by kind"),
	)
	if err != nil {
		h, _ = nm.Float64Histogram("gotd.downloader.request.duration")
	}
	o.duration = h

	for k, kind := range requestKinds {
		o.kindAttrs[k] = attrOpt(attribute.String("kind", kind))
		for r, result := range requestResults {
			o.requestAttrs[k][r] = attrOpt(
				attribute.String("kind", kind),
				attribute.String("result", result),
			)
		}
	}
	return o
}

type observerKey struct{}

// with returns context carrying o, so that schemas can report events.
func (o *observer) with(ctx context.Context) context.Context {
	if o == nil {
		return ctx
	}
	return context.WithValue(ctx, observerKey{}, o)
}

// observerFrom returns observer of the download, nil if not set.
func observerFrom(ctx context.Context) *observer {
	o, _ := ctx.Value(observerKey{}).(*observer)
	return o
}

func kindIndex(kind string) int {
	for i, k := range requestKinds {
		if k == kind {
			return i
		}
	}
	return 0
}

func requestResult(ctx context.Context, err error) int {
	switch {
	case err == nil:
		return 0
	case ctx.Err() != nil:
		return 3
	}
	if _, ok := tgerr.AsFloodWait(err); ok {
		return 1
	}
	if isRetryableTimeout(ctx, err) {
		return 2
	}
	return 4
}

// request reports a part request.
func (o *observer) request(ctx context.Context, kind string, start time.Time, n int, err error) {
	if o == nil {
		return
	}
	k := kindIndex(kind)
	o.requests.Add(ctx, 1, o.requestAttrs[k][requestResult(ctx, err)])
	o.duration.Record(ctx, time.Since(start).Seconds(), o.kindAttrs[k])
	if n > 0 {
		o.bytes.Add(ctx, int64(n), o.kindAttrs[k])
	}
}

// download reports a finished download.
func (o *observer) download(ctx context.Context, err error) {
	if o == nil {
		return
	}
	result := resultOK
	if err != nil {
		result = resultError
		if ctx.Err() != nil {
			result = resultCanceled
		}
	}
	o.downloads.Add(ctx, 1, attrOpt(attribute.String("result", result)))
	if err != nil {
		o.log.Debug(ctx, "Download failed", log.Error(err))
	}
}

func (o *observer) resume(ctx context.Context, offset int64) {
	if o == nil {
		return
	}
	o.resumes.Add(ctx, 1)
	o.log.Debug(ctx, "Resume download", log.Int64("offset", offset))
}

func (o *observer) cdnRedirect(ctx context.Context, dc int) {
	if o == nil {
		return
	}
	o.cdnRedirects.Add(ctx, 1)
	o.log.Debug(ctx, "Redirected to CDN", log.Int("cdn_dc", dc))
}

func (o *observer) cdnReupload(ctx context.Context, dc int) {
	if o == nil {
		return
	}
	o.cdnReuploads.Add(ctx, 1)
	o.log.Debug(ctx, "CDN file reupload", log.Int("cdn_dc", dc))
}

func (o *observer) hashMismatch(ctx context.Context, kind string, offset int64) {
	if o == nil {
		return
	}
	o.hashMismatches.Add(ctx, 1, attrOpt(attribute.String("kind", kind)))
	o.log.Warn(ctx, "Hash mismatch", log.String("kind", kind), log.Int64("offset", offset))
}

func (o *observer) refresh(ctx context.Context, result string, err error) {
	if o == nil {
		return
	}
	o.refreshes.Add(ctx, 1, attrOpt(attribute.String("result", result)))
	if result == refreshOK {
		o.log.Debug(ctx, "File reference refreshed")
		return
	}
	o.log.Warn(ctx, "File reference refresh failed", log.String("result", result), log.Error(err))
}

func (o *observer) backoff(ctx context.Context, err error) {
	if o == nil {
		return
	}
	reason := resultTimeout
	if _, ok := tgerr.AsFloodWait(err); ok {
		reason = resultFloodWait
	}
	o.backoffs.Add(ctx, 1, attrOpt(attribute.String("reason", reason)))
	o.log.Debug(ctx, "Scheduler backoff", log.String("reason", reason))
}

func (o *observer) verified(ctx context.Context, verified, size int64, err error) {
	if o == nil {
		return
	}
	result := verifyPartial
	switch {
	case err != nil:
		result = verifyError
	case verified >= size:
		result = verifyComplete
	case verified == 0:
		result = verifyNone
	}
	o.verify.Add(ctx, 1, attrOpt(attribute.String("result", result)))
	o.log.Debug(ctx, "Verified local file",
		log.String("result", result),
		log.Int64("verified", verified),
		log.Int64("size", size),
	)
}

func (o *observer) retry(e RetryEvent) {
	if o == nil {
		return
	}
	ctx := context.Background()
	o.retries.Add(ctx, 1, attrOpt(attribute.String("operation", e.Operation)))
	o.log.Debug(ctx, "Retry",
		log.String("operation", e.Operation),
		log.Int("attempt", e.Attempt),
		log.Error(e.Err),
	)
}

func (o *observer) exhausted(ctx context.Context, op string, err error) {
	if o == nil {
		return
	}
	o.retriesExhausted.Add(ctx, 1, attrOpt(attribute.String("operation", op)))
	o.log.Warn(ctx, "Retry limit reached", log.String("operation", op), log.Error(err))
}

// wrapRetry returns RetryHandler that reports retries before calling h.
func (o *observer) wrapRetry(h RetryHandler) RetryHandler {
	if o == nil {
		return h
	}
	return func(e RetryEvent) {
		o.retry(e)
		if h != nil {
			h(e)
		}
	}
}

// WithMeterProvider sets meter provider for downloader metrics. Metrics are
// not collected by default.
func (d *Downloader) WithMeterProvider(mp metric.MeterProvider) *Downloader {
	d.meterProvider = mp
	d.obs = newObserver(d.meterProvider, d.logger)
	return d
}

// WithLogger sets logger for downloader debug logs. No logs by default.
func (d *Downloader) WithLogger(l log.Logger) *Downloader {
	d.logger = l
	d.obs = newObserver(d.meterProvider, d.logger)
	return d
}

// Instrument registers observable instruments of Scheduler state on m:
// gotd.downloader.scheduler.window, .in_flight and .waiting gauges and the
// .decreases counter, with given attributes (e.g. DC).
//
// Unregister the returned registration when the Scheduler is not used
// anymore.
func (s *Scheduler) Instrument(m metric.Meter, attrs ...attribute.KeyValue) (metric.Registration, error) {
	window, err := m.Int64ObservableGauge("gotd.downloader.scheduler.window",
		metric.WithUnit("{request}"),
		metric.WithDescription("Scheduler limit of in-flight part requests"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "window")
	}
	inFlight, err := m.Int64ObservableGauge("gotd.downloader.scheduler.in_flight",
		metric.WithUnit("{request}"),
		metric.WithDescription("Scheduler in-flight part requests"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "in flight")
	}
	waiting, err := m.Int64ObservableGauge("gotd.downloader.scheduler.waiting",
		metric.WithUnit("{worker}"),
		metric.WithDescription("Workers waiting for a Scheduler slot"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "waiting")
	}
	decreases, err := m.Int64ObservableCounter("gotd.downloader.scheduler.decreases",
		metric.WithUnit("{decrease}"),
		metric.WithDescription("Scheduler window decreases"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "decreases")
	}

	opt := attrOpt(attrs...)
	reg, err := m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		snap := s.Snapshot()
		o.ObserveInt64(window, int64(snap.Window), opt)
		o.ObserveInt64(inFlight, int64(snap.InFlight), opt)
		o.ObserveInt64(waiting, int64(snap.Waiting), opt)
		o.ObserveInt64(decreases, snap.Decreases, opt)
		return nil
	}, window, inFlight, waiting, decreases)
	if err != nil {
		return nil, errors.Wrap(err, "register callback")
	}
	return reg, nil
}
