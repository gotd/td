package telegram

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/tg"
)

// histogramSum returns the sum of histogram data points of name with given
// attributes.
func (m *testMetrics) histogramSum(name string, kv ...attribute.KeyValue) float64 {
	m.t.Helper()
	var total float64
	for _, metric := range m.collect() {
		if metric.Name != name {
			continue
		}
		data, ok := metric.Data.(metricdata.Histogram[float64])
		require.True(m.t, ok, "%s is %T", name, metric.Data)
		for _, dp := range data.DataPoints {
			if hasAttrs(dp.Attributes, kv) {
				total += dp.Sum
			}
		}
	}
	return total
}

// recordLogger keeps records of one message.
type recordLogger struct {
	msg string

	mux     sync.Mutex
	records [][]log.Attr
}

func (*recordLogger) Enabled(context.Context, log.Level) bool { return true }

func (l *recordLogger) Log(_ context.Context, _ log.Level, msg string, attrs ...log.Attr) {
	if msg != l.msg {
		return
	}
	l.mux.Lock()
	defer l.mux.Unlock()
	l.records = append(l.records, append([]log.Attr(nil), attrs...))
}

// testServerMsgID returns a server msg_id created about at, and its exact
// time: the generator adds jitter to ids.
func testServerMsgID(at time.Time) (int64, time.Time) {
	id := proto.NewMessageIDGen(func() time.Time { return at }).New(proto.MessageFromServer)
	return id, proto.MessageID(id).Time()
}

func TestClientUpdateTiming(t *testing.T) {
	const (
		receiveDelay = 40 * time.Millisecond
		process      = 2 * time.Millisecond
		handle       = 10 * time.Millisecond
		// Seconds, to compare float sums.
		delta = 1e-6
	)
	date := time.Unix(1_700_000_000, 0)
	msgID, created := testServerMsgID(date.Add(2500 * time.Millisecond))
	var (
		serverDelay = created.Sub(date)
		received    = created.Add(receiveDelay)
		clock       = neo.NewTime(received.Add(process))
		logger      = &recordLogger{msg: "Update timing"}
	)

	c := newTestClient(nil)
	c.clock = clock
	c.log = log.For(logger)
	c.updateHandler = UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
		clock.Travel(handle)
		return nil
	})
	m := setTestMetrics(t, c)

	updates := &tg.Updates{
		Updates: []tg.UpdateClass{
			&tg.UpdateNewChannelMessage{
				Message: &tg.Message{
					ID:      1,
					PeerID:  &tg.PeerChannel{ChannelID: 4242},
					Date:    int(date.Unix()),
					Message: "secret text",
				},
			},
			// An edit has the date of the original message, so it is not a
			// delay of this update.
			&tg.UpdateEditChannelMessage{
				Message: &tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 4242}, Date: int(date.Add(-time.Hour).Unix())},
			},
			&tg.UpdateNewMessage{
				Message: &tg.Message{ID: 2, PeerID: &tg.PeerUser{UserID: 77}, Date: int(date.Add(time.Second).Unix())},
			},
		},
	}
	var b bin.Buffer
	require.NoError(t, b.Encode(updates))

	require.NoError(t, c.asHandler().(mtproto.MessageInfoHandler).OnMessageInfo(
		mtproto.MessageInfo{ID: msgID, Received: received}, &b,
	))

	const (
		serverName  = "gotd.client.updates.server.delay"
		receiveName = "gotd.client.updates.receive.delay"
		processName = "gotd.client.updates.process.duration"
		handleName  = "gotd.client.updates.handle.duration"
	)
	require.EqualValues(t, 1, m.value(serverName, strAttr("kind", "channel")))
	require.EqualValues(t, 1, m.value(serverName, strAttr("kind", "other")))
	require.InDelta(t, serverDelay.Seconds(), m.histogramSum(serverName, strAttr("kind", "channel")), delta)
	require.InDelta(t, (serverDelay - time.Second).Seconds(), m.histogramSum(serverName, strAttr("kind", "other")), delta)
	require.InDelta(t, receiveDelay.Seconds(), m.histogramSum(receiveName), delta)
	require.InDelta(t, process.Seconds(), m.histogramSum(processName), delta)
	require.InDelta(t, handle.Seconds(), m.histogramSum(handleName, strAttr("result", "ok")), delta)
	require.Zero(t, m.value(handleName, strAttr("result", "error")))

	// The record has durations and counts only: no peers, ids or text.
	require.Len(t, logger.records, 1)
	got := map[string]string{}
	for _, a := range logger.records[0] {
		got[a.Key] = a.Value.String()
	}
	require.Len(t, got, 6)
	for _, key := range []string{"receive_delay", "process", "handle", "server_delay"} {
		require.Contains(t, got, key)
	}
	require.Equal(t, "2", got["messages"])
	require.Equal(t, "1", got["channel_messages"])
}

// The server time offset removes the clock difference from the receive
// delay: here the local clock is 300ms behind and the message took 20ms more
// than the fastest one.
func TestClientUpdateTimingServerOffset(t *testing.T) {
	msgID, created := testServerMsgID(time.Unix(1_700_000_000, 0))
	received := created.Add(-280 * time.Millisecond)

	c := newTestClient(nil)
	c.clock = neo.NewTime(received)
	m := setTestMetrics(t, c)

	var b bin.Buffer
	require.NoError(t, b.Encode(&tg.UpdatesTooLong{}))
	require.NoError(t, c.handleUpdatesInfo(mtproto.MessageInfo{
		ID: msgID, Received: received, ServerTimeOffset: 300 * time.Millisecond,
	}, &b))

	require.InDelta(t, 0.02, m.histogramSum("gotd.client.updates.receive.delay"), 1e-6)
}

// Without the offset a local clock behind the server makes the receive delay
// negative. It must be recorded as is: a clamp would hide the skew.
func TestClientUpdateTimingClockSkew(t *testing.T) {
	msgID, created := testServerMsgID(time.Unix(1_700_000_000, 0))
	received := created.Add(-300 * time.Millisecond)

	c := newTestClient(nil)
	c.clock = neo.NewTime(received)
	m := setTestMetrics(t, c)

	var b bin.Buffer
	require.NoError(t, b.Encode(&tg.UpdatesTooLong{}))
	require.NoError(t, c.handleUpdatesInfo(mtproto.MessageInfo{ID: msgID, Received: received}, &b))

	require.EqualValues(t, 1, m.value("gotd.client.updates.receive.delay"))
	require.InDelta(t, -0.3, m.histogramSum("gotd.client.updates.receive.delay"), 1e-6)
	require.Zero(t, m.value("gotd.client.updates.server.delay"))
}

// Without the receive time (not from a connection) nothing is measured.
func TestClientUpdateTimingNoInfo(t *testing.T) {
	c := newTestClient(nil)
	m := setTestMetrics(t, c)

	var b bin.Buffer
	require.NoError(t, b.Encode(&tg.UpdatesTooLong{}))
	require.NoError(t, c.handleUpdates(&b))
	require.Zero(t, m.value("gotd.client.updates.handle.duration"))
}
