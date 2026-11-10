package telegram

import (
	"context"
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

	"github.com/gotd/td/bin"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// testMetrics collects metrics recorded by a client.
type testMetrics struct {
	t      *testing.T
	reader *sdkmetric.ManualReader
}

// setTestMetrics replaces metrics of c with ones collected by the returned
// testMetrics.
func setTestMetrics(t *testing.T, c *Client) *testMetrics {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	m, err := newClientMetrics(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	require.NoError(t, err)
	c.metrics = m
	return &testMetrics{t: t, reader: reader}
}

func (m *testMetrics) collect() []metricdata.Metrics {
	m.t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(m.t, m.reader.Collect(context.Background(), &rm))
	var out []metricdata.Metrics
	for _, sm := range rm.ScopeMetrics {
		out = append(out, sm.Metrics...)
	}
	return out
}

func hasAttrs(set attribute.Set, kv []attribute.KeyValue) bool {
	for _, want := range kv {
		got, ok := set.Value(want.Key)
		if !ok || got != want.Value {
			return false
		}
	}
	return true
}

// value returns the sum of counter or gauge data points of name with given
// attributes, or the count of histogram data points.
func (m *testMetrics) value(name string, kv ...attribute.KeyValue) int64 {
	m.t.Helper()
	var total int64
	for _, metric := range m.collect() {
		if metric.Name != name {
			continue
		}
		switch data := metric.Data.(type) {
		case metricdata.Sum[int64]:
			for _, dp := range data.DataPoints {
				if hasAttrs(dp.Attributes, kv) {
					total += dp.Value
				}
			}
		case metricdata.Gauge[int64]:
			for _, dp := range data.DataPoints {
				if hasAttrs(dp.Attributes, kv) {
					total += dp.Value
				}
			}
		case metricdata.Histogram[float64]:
			for _, dp := range data.DataPoints {
				if hasAttrs(dp.Attributes, kv) {
					total += int64(dp.Count)
				}
			}
		default:
			m.t.Fatalf("unexpected data %T of %s", data, name)
		}
	}
	return total
}

func dcAttr(dc int) attribute.KeyValue { return attribute.Int("dc", dc) }

func strAttr(k, v string) attribute.KeyValue { return attribute.String(k, v) }

func TestNewClientMeterProvider(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	c := NewClient(1, "hash", Options{
		Logger:        log.Nop,
		MeterProvider: sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
	})
	m := &testMetrics{t: t, reader: reader}
	c.notifyConnectionState(ConnectionStateReady, 2)
	c.notifyConnDead(ConnDeadEvent{DC: 4, Kind: ConnKindCDN, Err: errors.New("dead")})

	require.EqualValues(t, 1, m.value("gotd.client.connection.state", dcAttr(2), strAttr("state", "ready")))
	require.EqualValues(t, 1, m.value("gotd.client.connection.dead", dcAttr(4), strAttr("kind", "cdn")))

	// No provider: noop metrics.
	require.NotNil(t, NewClient(1, "hash", Options{Logger: log.Nop}).meters())
	require.NotNil(t, (&Client{}).meters(), "client without NewClient")
}

func TestClientMetricsMigration(t *testing.T) {
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.UsersGetUsersRequest:
			return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
		case *tg.AuthLogOutRequest:
			if dc == 2 {
				return nil, tgerr.New(303, "USER_MIGRATE_10")
			}
			return &tg.BoolTrue{}, nil
		default:
			return nil, errors.Errorf("unexpected body %T", body)
		}
	})
	m := setTestMetrics(t, client)

	require.NoError(t, client.Run(context.Background(), func(ctx context.Context) error {
		var result tg.BoolTrue
		if err := client.Invoke(ctx, &tg.AuthLogOutRequest{}, &result); err != nil {
			return err
		}
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if m.value("gotd.client.connection.state", dcAttr(10), strAttr("state", "ready")) > 0 {
				return nil
			}
			time.Sleep(time.Millisecond)
		}
		return errors.New("no ready state of new primary connection")
	}))

	for _, tt := range []struct {
		dc    int
		state string
	}{
		{2, "connecting"},
		{2, "ready"},
		{2, "disconnected"},
		{10, "connecting"},
		{10, "ready"},
	} {
		require.EqualValues(t, 1,
			m.value("gotd.client.connection.state", dcAttr(tt.dc), strAttr("state", tt.state)),
			"%d %s", tt.dc, tt.state,
		)
	}
	require.EqualValues(t, 1, m.value("gotd.client.migrations", dcAttr(10)))
	require.EqualValues(t, 1, m.value("gotd.client.connection.restarts", strAttr("reason", "migration")))
	require.EqualValues(t, 1, m.value("gotd.client.connection.reconnects", dcAttr(2)))
	require.EqualValues(t, 1, m.value("gotd.client.invoke.duration", strAttr("result", "ok")),
		"auth.logOut is measured once, with migration",
	)
}

func TestClientMetricsRequestTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	good := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		return &tg.NearestDC{ThisDC: 2}, nil
	})
	c := newTestClient(nil)
	c.conn = &hangConn{}
	c.ready.Signal()
	c.requestTimeout = func(input bin.Encoder) time.Duration { return 50 * time.Millisecond }
	fakeReconnect(t, c, good.conn)
	m := setTestMetrics(t, c)

	_, err := c.API().HelpGetNearestDC(ctx)
	require.NoError(t, err)

	require.EqualValues(t, 1, m.value("gotd.client.request.timeouts",
		strAttr("result", "resent"), strAttr("method", "help.getNearestDc"),
	))
	require.Zero(t, m.value("gotd.client.request.timeouts", strAttr("result", "failed")))
	require.EqualValues(t, 1, m.value("gotd.client.connection.restarts", strAttr("reason", "timeout")))
	require.EqualValues(t, 1, m.value("gotd.client.invoke.duration", strAttr("result", "ok")))
}

func TestClientMetricsPools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{},
		fail:    map[int]int{},
	}
	var (
		mux        sync.Mutex
		exported   bool
		authorized bool
	)
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.UploadGetFileRequest:
			if dc == 3 {
				return nil, tgerr.New(303, "FILE_MIGRATE_4")
			}
			return &tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: []byte{byte(dc)}}, nil
		case *tg.AuthExportAuthorizationRequest:
			r := body.(*tg.AuthExportAuthorizationRequest)
			mux.Lock()
			first := r.DCID == 5 && !exported
			if r.DCID == 5 {
				exported = true
			}
			mux.Unlock()
			if first {
				// Not signed in yet: transfer fails, pool is still returned.
				_, _ = s.handle(id, dc, body)
				return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
			}
		case *tg.AuthImportAuthorizationRequest:
			if dc == 5 {
				mux.Lock()
				authorized = true
				mux.Unlock()
			}
		case *tg.HelpGetNearestDCRequest:
			mux.Lock()
			ok := authorized
			mux.Unlock()
			if dc == 5 && !ok {
				return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
			}
		}
		return s.handle(id, dc, body)
	})
	client.onTransfer = noopOnTransfer
	cfg := client.cfg.Load()
	for _, dc := range []int{3, 4, 5} {
		cfg.DCOptions = append(cfg.DCOptions, tg.DCOption{ID: dc, IPAddress: "dc"})
	}
	client.cfg.Store(cfg)
	m := setTestMetrics(t, client)
	unregister := client.registerPoolsGauge()
	defer unregister()

	// FILE_MIGRATE from DC 3 is followed to DC 4.
	inv, err := client.FileDC(ctx, 3)
	require.NoError(t, err)
	_, err = tg.NewClient(inv).UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: &tg.InputDocumentFileLocation{ID: 1},
		Limit:    1024,
	})
	require.NoError(t, err)

	// Migrate pool with failed transfer is recreated.
	require.NoError(t, invokeSubNearest(ctx, client, 5))

	require.EqualValues(t, 1, m.value("gotd.client.file.migrations", dcAttr(4), strAttr("source", "file")))
	for _, dc := range []int{3, 4} {
		require.EqualValues(t, 1, m.value("gotd.client.pool.created",
			dcAttr(dc), strAttr("kind", "file"), strAttr("result", "ok"),
		), "file pool of DC %d", dc)
	}
	require.EqualValues(t, 2, m.value("gotd.client.pool.created",
		dcAttr(5), strAttr("kind", "migrate"), strAttr("result", "ok"),
	))
	require.EqualValues(t, 1, m.value("gotd.client.pool.evicted",
		dcAttr(5), strAttr("kind", "migrate"), strAttr("reason", "unauthorized"),
	))
	require.EqualValues(t, 4, m.value("gotd.client.pool.create.duration", strAttr("result", "ok")))
	require.EqualValues(t, 2, m.value("gotd.client.pools", strAttr("kind", "file")))
	require.EqualValues(t, 1, m.value("gotd.client.pools", strAttr("kind", "migrate")))
	require.EqualValues(t, 0, m.value("gotd.client.pools", strAttr("kind", "cdn")))
}

func TestClientMetricsPoolCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release := make(chan struct{})
	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{4: release},
		fail:    map[int]int{},
	}
	client := newSubConnsTestClient(t, s)
	clientCtx, closeClient := context.WithCancel(context.Background())
	client.ctx = clientCtx
	m := setTestMetrics(t, client)

	errs := make(chan error, 1)
	go func() { errs <- invokeSubNearest(ctx, client, 4) }()
	require.Eventually(t, func() bool { return s.exportCount(4) == 1 }, 5*time.Second, time.Millisecond)

	// Client shutdown during creation is not a failure of the DC.
	closeClient()
	close(release)
	require.Error(t, <-errs)

	require.Eventually(t, func() bool {
		return m.value("gotd.client.pool.created",
			dcAttr(4), strAttr("kind", "migrate"), strAttr("result", "canceled"),
		) == 1
	}, 5*time.Second, time.Millisecond)
	require.Zero(t, m.value("gotd.client.pool.created", strAttr("result", "error")))
}

func TestClientMetricsRestartAbandoned(t *testing.T) {
	c := newTestClient(nil)
	m := setTestMetrics(t, c)

	// Nobody receives the restart: it is abandoned on cancel and not counted.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, c.ensureRestart(ctx, restartTimeout), context.Canceled)
	require.Zero(t, m.value("gotd.client.connection.restarts"))
}

func TestClientMetricsAuthLost(t *testing.T) {
	ctx := context.Background()
	var fail atomic.Pointer[tgerr.Error]
	client := newTestClient(authFlowHandler(&fail, nil, nil))
	m := setTestMetrics(t, client)
	api := client.API()

	// Not authorized yet: not lost.
	fail.Store(tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
	_, _ = api.HelpGetConfig(ctx)
	require.Zero(t, m.value("gotd.client.auth.lost"))

	fail.Store(nil)
	_, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	require.NoError(t, err)

	// Counted once per lost authorization, without OnAuthLost too.
	fail.Store(tgerr.New(401, "SESSION_REVOKED"))
	_, _ = api.HelpGetConfig(ctx)
	_, _ = api.HelpGetConfig(ctx)
	require.EqualValues(t, 1, m.value("gotd.client.auth.lost",
		strAttr("reason", "SESSION_REVOKED"), strAttr("kind", "revoked"),
	))

	// Restrictions are counted on every error, authorized or not.
	fail.Store(tgerr.New(400, "PHONE_NUMBER_BANNED"))
	_, _ = api.HelpGetConfig(ctx)
	require.EqualValues(t, 1, m.value("gotd.client.account.restricted",
		strAttr("reason", "PHONE_NUMBER_BANNED"),
	))

	require.EqualValues(t, 3, m.value("gotd.client.invoke.duration",
		strAttr("result", "rpc_error"), attribute.Int("code", 401),
	))
	require.EqualValues(t, 1, m.value("gotd.client.invoke.duration",
		strAttr("result", "rpc_error"), attribute.Int("code", 400),
	))
	require.EqualValues(t, 1, m.value("gotd.client.invoke.duration", strAttr("result", "ok")))
}

func TestClientMetricsInvokeResult(t *testing.T) {
	m, err := newClientMetrics(nil)
	require.NoError(t, err)
	ctx := context.Background()
	canceled, cancel := context.WithCancel(ctx)
	cancel()

	require.Equal(t, m.invokeOK, m.invokeResult(ctx, nil))
	require.Equal(t, m.invokeCanceled, m.invokeResult(canceled, errors.New("conn closed")))
	require.Equal(t, m.invokeCanceled, m.invokeResult(ctx, errors.Wrap(context.DeadlineExceeded, "wait")))
	require.Equal(t, m.invokeError, m.invokeResult(ctx, errors.New("conn closed")))

	// RPC error attributes are cached per code.
	flood := m.invokeResult(ctx, tgerr.New(420, "FLOOD_WAIT_3"))
	require.Equal(t, flood, m.invokeResult(ctx, tgerr.New(420, "FLOOD_PREMIUM_WAIT_3")))
	require.NotEqual(t, flood, m.invokeResult(ctx, tgerr.New(400, "PEER_ID_INVALID")))
}

func TestMeasuredInvokerZeroAlloc(t *testing.T) {
	c := &Client{}
	inv := measuredInvoker{client: c, next: InvokeFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return nil
	})}
	ctx := context.Background()
	testutil.ZeroAlloc(t, func() {
		_ = inv.Invoke(ctx, nil, nil)
	})
}

func BenchmarkMeasuredInvoker(b *testing.B) {
	c := &Client{}
	inv := measuredInvoker{client: c, next: InvokeFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return nil
	})}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		_ = inv.Invoke(ctx, nil, nil)
	}
}
