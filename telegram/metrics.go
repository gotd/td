package telegram

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/gotd/log"

	"github.com/gotd/td/tgerr"
)

// meterName is the name of the meter of Client.
const meterName = "github.com/gotd/td/telegram"

// Values of metric attributes.
const (
	poolKindMigrate = "migrate" // FILE_MIGRATE/STATS_MIGRATE pools (invokeSub).
	poolKindFile    = "file"    // FileDC pools.
	poolKindCDN     = "cdn"     // CDN pools.

	restartTimeout   = "timeout"
	restartMigration = "migration"
)

// durationBuckets are histogram buckets in seconds, from a fast RPC to a slow
// pool creation.
var durationBuckets = metric.WithExplicitBucketBoundaries( // nolint:gochecknoglobals
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60,
)

// clientMetrics are OpenTelemetry instruments of Client.
//
// Attributes are low-cardinality only: DC IDs, kinds, states, results, and
// TL method names of timed out requests. No peers or message content.
type clientMetrics struct {
	connState      metric.Int64Counter
	connDead       metric.Int64Counter
	reconnects     metric.Int64Counter
	restarts       metric.Int64Counter
	migrations     metric.Int64Counter
	authLost       metric.Int64Counter
	requestTimeout metric.Int64Counter
	poolCreated    metric.Int64Counter
	poolDuration   metric.Float64Histogram
	poolEvicted    metric.Int64Counter
	fileMigrations metric.Int64Counter
	invokeDuration metric.Float64Histogram
	pools          metric.Int64ObservableGauge

	meter metric.Meter

	// Pre-built attributes of hot paths, as slices to pass without
	// allocation.
	invokeOK       []metric.RecordOption
	invokeCanceled []metric.RecordOption
	invokeError    []metric.RecordOption
	// rpcCodes caches attributes of RPC error codes: int -> []metric.RecordOption.
	rpcCodes sync.Map
}

// newClientMetrics creates instruments of Client. A nil provider means no
// metrics.
func newClientMetrics(provider metric.MeterProvider) (*clientMetrics, error) {
	if provider == nil {
		provider = noop.NewMeterProvider()
	}
	meter := provider.Meter(meterName)
	m := &clientMetrics{
		meter:          meter,
		invokeOK:       resultOptions(attribute.String("result", "ok")),
		invokeCanceled: resultOptions(attribute.String("result", "canceled")),
		invokeError:    resultOptions(attribute.String("result", "error")),
	}

	var err error
	counter := func(name, unit, desc string) metric.Int64Counter {
		c, cErr := meter.Int64Counter(name, metric.WithUnit(unit), metric.WithDescription(desc))
		err = errors.Join(err, cErr)
		return c
	}
	histogram := func(name, desc string) metric.Float64Histogram {
		h, hErr := meter.Float64Histogram(name,
			metric.WithUnit("s"), metric.WithDescription(desc), durationBuckets,
		)
		err = errors.Join(err, hErr)
		return h
	}

	m.connState = counter("gotd.client.connection.state", "{event}",
		"Primary connection state changes, by DC and state.")
	m.connDead = counter("gotd.client.connection.dead", "{connection}",
		"Dead connections, by DC and kind (primary, data, cdn).")
	m.reconnects = counter("gotd.client.connection.reconnects", "{reconnect}",
		"Primary connection reconnects after the connection stopped, by DC.")
	m.restarts = counter("gotd.client.connection.restarts", "{restart}",
		"Primary connection restarts requested by the client, by reason (timeout, migration). "+
			"Each restart is also counted in gotd.client.connection.reconnects.")
	m.migrations = counter("gotd.client.migrations", "{migration}",
		"Primary DC migrations, by target DC.")
	m.authLost = counter("gotd.client.auth.lost", "{event}",
		"Lost authorizations of an authorized client, by RPC error type.")
	m.requestTimeout = counter("gotd.client.request.timeouts", "{request}",
		"Requests without response in Options.RequestTimeout, by result (resent, failed) and method.")
	m.poolCreated = counter("gotd.client.pool.created", "{pool}",
		"Cached DC pools created, by DC, kind (file, migrate) and result.")
	m.poolDuration = histogram("gotd.client.pool.create.duration",
		"Duration of cached DC pool creation, with authorization transfer, by kind and result.")
	m.poolEvicted = counter("gotd.client.pool.evicted", "{pool}",
		"Cached DC pools dropped after an unauthorized error, by DC and kind.")
	m.fileMigrations = counter("gotd.client.file.migrations", "{redirect}",
		"FILE_MIGRATE redirects followed, by target DC and source (primary, file).")
	m.invokeDuration = histogram("gotd.client.invoke.duration",
		"Duration of Client.Invoke, with middlewares, by result (ok, rpc_error with code, canceled, error).")

	pools, pErr := meter.Int64ObservableGauge("gotd.client.pools",
		metric.WithUnit("{pool}"),
		metric.WithDescription("Open cached pools, by kind (file, migrate, cdn)."),
	)
	err = errors.Join(err, pErr)
	m.pools = pools

	if err != nil {
		return nil, err
	}
	return m, nil
}

var (
	noopMetricsOnce sync.Once      // nolint:gochecknoglobals
	noopMetrics     *clientMetrics // nolint:gochecknoglobals
)

// noopClientMetrics returns shared metrics that record nothing.
func noopClientMetrics() *clientMetrics {
	noopMetricsOnce.Do(func() {
		m, err := newClientMetrics(noop.NewMeterProvider())
		if err != nil {
			// Noop instruments never fail.
			panic(err)
		}
		noopMetrics = m
	})
	return noopMetrics
}

// meters returns instruments of the client, or noop ones for a client
// created without NewClient (tests).
func (c *Client) meters() *clientMetrics {
	if c.metrics == nil {
		return noopClientMetrics()
	}
	return c.metrics
}

// resultOptions returns record options with given attributes.
func resultOptions(kv ...attribute.KeyValue) []metric.RecordOption {
	return []metric.RecordOption{metric.WithAttributeSet(attribute.NewSet(kv...))}
}

// invokeResult returns attributes of an Invoke result.
func (m *clientMetrics) invokeResult(ctx context.Context, err error) []metric.RecordOption {
	if err == nil {
		return m.invokeOK
	}
	if rpcErr, ok := tgerr.As(err); ok {
		if v, ok := m.rpcCodes.Load(rpcErr.Code); ok {
			return v.([]metric.RecordOption)
		}
		opts := resultOptions(
			attribute.String("result", "rpc_error"),
			attribute.Int("code", rpcErr.Code),
		)
		m.rpcCodes.Store(rpcErr.Code, opts)
		return opts
	}
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return m.invokeCanceled
	}
	return m.invokeError
}

// dcAttrs returns attributes with DC ID.
func dcAttrs(dc int, kv ...attribute.KeyValue) metric.MeasurementOption {
	return metric.WithAttributes(append(kv, attribute.Int("dc", dc))...)
}

// now returns current time of the client clock.
func (c *Client) now() time.Time {
	if c.clock == nil {
		return time.Now()
	}
	return c.clock.Now()
}

// registerPoolsGauge starts reporting open pools until the returned function
// is called.
func (c *Client) registerPoolsGauge() func() {
	m := c.meters()
	attrFile := metric.WithAttributeSet(attribute.NewSet(attribute.String("kind", poolKindFile)))
	attrMigrate := metric.WithAttributeSet(attribute.NewSet(attribute.String("kind", poolKindMigrate)))
	attrCDN := metric.WithAttributeSet(attribute.NewSet(attribute.String("kind", poolKindCDN)))
	reg, err := m.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveInt64(m.pools, int64(c.filePools.len()), attrFile)
		o.ObserveInt64(m.pools, int64(c.subConns.len()), attrMigrate)
		o.ObserveInt64(m.pools, int64(c.cdnPools.count()), attrCDN)
		return nil
	}, m.pools)
	if err != nil {
		c.log.Warn(context.Background(), "Register pools gauge", log.Error(err))
		return func() {}
	}
	return func() { _ = reg.Unregister() }
}
