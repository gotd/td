package mtproto

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
)

// sentPings returns the ping_delay_disconnect requests traced so far.
func sentPings(t *testing.T, rec *wireRecorder) []mt.PingDelayDisconnectRequest {
	t.Helper()
	rec.mux.Lock()
	defer rec.mux.Unlock()
	var out []mt.PingDelayDisconnectRequest
	for _, m := range rec.msgs {
		if m.m.Direction != WireOut || len(m.body) < 4 ||
			binary.LittleEndian.Uint32(m.body) != mt.PingDelayDisconnectRequestTypeID {
			continue
		}
		var req mt.PingDelayDisconnectRequest
		require.NoError(t, req.Decode(&bin.Buffer{Buf: append([]byte(nil), m.body...)}))
		out = append(out, req)
	}
	return out
}

func runPingLoop(t *testing.T, interval, timeout time.Duration) (*neo.Time, *Conn, *wireRecorder, func()) {
	t.Helper()
	rec := &wireRecorder{}
	c := newWireTestConn(t, rec, -1)
	clk := neo.NewTime(testutil.Date())
	c.clock = clk
	c.conn = &constantConn{}
	c.ping = map[int64]chan struct{}{}
	c.messageID = proto.NewMessageIDGen(clk.Now)
	c.pingInterval = interval
	c.pingTimeout = timeout

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	armed := clk.Observe()
	go func() {
		defer close(done)
		_ = c.pingLoop(ctx)
	}()
	<-armed // first ping timer is set
	return clk, c, rec, func() { cancel(); <-done }
}

// pong answers the ping and waits until the loop sets the next ping timer.
func pong(t *testing.T, clk *neo.Time, c *Conn, req mt.PingDelayDisconnectRequest) {
	t.Helper()
	rearmed := clk.Observe()
	var b bin.Buffer
	require.NoError(t, (&mt.Pong{PingID: req.PingID}).Encode(&b))
	require.NoError(t, c.handlePong(&b))
	<-rearmed
}

func waitPings(t *testing.T, rec *wireRecorder, n int) []mt.PingDelayDisconnectRequest {
	t.Helper()
	var pings []mt.PingDelayDisconnectRequest
	require.Eventually(t, func() bool {
		pings = sentPings(t, rec)
		return len(pings) >= n
	}, 5*time.Second, time.Millisecond)
	return pings
}

// The default cadence is Telegram Desktop's: ping_delay_disconnect(60) at
// least 30 s (kPingSendAfter) after the previous ping.
func TestPingLoopDefaultCadence(t *testing.T) {
	clk, c, rec, stop := runPingLoop(t, 0, 15*time.Second)
	defer stop()

	// Zero random source: the interval is exactly pingSendAfter.
	clk.Travel(pingSendAfter - time.Second)
	time.Sleep(10 * time.Millisecond)
	require.Empty(t, sentPings(t, rec), "no ping before 30 s")

	clk.Travel(time.Second)
	pings := waitPings(t, rec, 1)
	require.Equal(t, pingDelayDisconnect, pings[0].DisconnectDelay)
	pong(t, clk, c, pings[0])

	clk.Travel(pingSendAfter)
	pings = waitPings(t, rec, 2)
	require.Equal(t, 60, pings[1].DisconnectDelay)
}

func TestNextPingDefaultRange(t *testing.T) {
	c := newWireTestConn(t, nil, -1)
	d, delay := c.nextPing()
	require.Equal(t, pingSendAfter, d, "zero random source")
	require.Equal(t, 60, delay)

	c.rand = testutil.Rand([]byte{1})
	seen := map[bool]bool{}
	for i := 0; i < 100; i++ {
		d, delay := c.nextPing()
		require.GreaterOrEqual(t, d, pingSendAfter)
		require.LessOrEqual(t, d, pingSendAfterForce)
		require.Equal(t, 60, delay)
		seen[d > pingSendAfter+7*time.Second] = true
	}
	require.Len(t, seen, 2, "interval is jittered over the range")
}

// An explicit PingInterval keeps a fixed interval with
// disconnect_delay = PingInterval + PingTimeout.
func TestPingLoopFixedInterval(t *testing.T) {
	clk, c, rec, stop := runPingLoop(t, 10*time.Second, 5*time.Second)
	defer stop()

	clk.Travel(10 * time.Second)
	pings := waitPings(t, rec, 1)
	require.Equal(t, 15, pings[0].DisconnectDelay)
	pong(t, clk, c, pings[0])

	clk.Travel(10 * time.Second)
	pings = waitPings(t, rec, 2)
	require.Equal(t, 15, pings[1].DisconnectDelay)
}
