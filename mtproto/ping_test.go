package mtproto

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/transport"
)

// fakeTransport records encrypted pings and optionally reports bytes read.
type fakeTransport struct {
	key  crypto.AuthKey
	rand io.Reader

	bytes   atomic.Uint64
	samples atomic.Uint64

	mu        sync.Mutex
	sendErr   error
	decodeErr error
	sends     []mt.PingDelayDisconnectRequest
	onSend    func(mt.PingDelayDisconnectRequest)
}

var _ transport.ReadCounter = (*fakeTransport)(nil)

func (f *fakeTransport) BytesRead() uint64 {
	f.samples.Add(1)
	return f.bytes.Load()
}

func (f *fakeTransport) Send(_ context.Context, b *bin.Buffer) error {
	f.mu.Lock()
	err := f.sendErr
	f.mu.Unlock()
	if err != nil {
		return err
	}

	buf := &bin.Buffer{Buf: append([]byte(nil), b.Buf...)}
	msg, decErr := crypto.NewServerCipher(f.rand).DecryptFromBuffer(f.key, buf)
	if decErr != nil {
		f.mu.Lock()
		f.decodeErr = decErr
		f.mu.Unlock()
		return nil
	}
	var req mt.PingDelayDisconnectRequest
	if decErr = req.Decode(&bin.Buffer{Buf: msg.Data()}); decErr != nil {
		f.mu.Lock()
		f.decodeErr = decErr
		f.mu.Unlock()
		return nil
	}

	f.mu.Lock()
	f.sends = append(f.sends, req)
	fn := f.onSend
	f.mu.Unlock()
	if fn != nil {
		fn(req)
	}
	return nil
}

func (f *fakeTransport) Recv(context.Context, *bin.Buffer) error { return nil }
func (f *fakeTransport) Close() error                            { return nil }

func (f *fakeTransport) sent() []mt.PingDelayDisconnectRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]mt.PingDelayDisconnectRequest, len(f.sends))
	copy(out, f.sends)
	return out
}

type captureLog struct {
	mu   sync.Mutex
	msgs []string
}

func (l *captureLog) Enabled(context.Context, log.Level) bool { return true }
func (l *captureLog) Log(_ context.Context, _ log.Level, msg string, _ ...log.Attr) {
	l.mu.Lock()
	l.msgs = append(l.msgs, msg)
	l.mu.Unlock()
}

// readyClock reports when the first ticker has been fully created.
type readyClock struct {
	*neo.Time
	ready chan struct{}
	once  sync.Once
}

func (c *readyClock) Now() time.Time { return c.Time.Now() }

func (c *readyClock) Timer(d time.Duration) clock.Timer { return c.Time.Timer(d) }

func (c *readyClock) Ticker(d time.Duration) clock.Ticker {
	tk := c.Time.Ticker(d)
	c.once.Do(func() { close(c.ready) })
	return tk
}

func (l *captureLog) has(msg string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, m := range l.msgs {
		if m == msg {
			return true
		}
	}
	return false
}

func newPingTestConn(t *testing.T, clock *neo.Time, fc *fakeTransport, interval, timeout, maxWait time.Duration, aware bool, logger log.Logger) *Conn {
	t.Helper()
	var key crypto.Key
	_, err := io.ReadFull(rand.Reader, key[:])
	require.NoError(t, err)
	authKey := key.WithID()
	fc.key = authKey
	fc.rand = rand.Reader

	return &Conn{
		conn:              fc,
		clock:             clock,
		rand:              rand.Reader,
		cipher:            crypto.NewClientCipher(rand.Reader),
		log:               log.For(logger),
		messageID:         proto.NewMessageIDGen(clock.Now),
		authKey:           authKey,
		compressThreshold: -1,
		ping:              map[int64]chan struct{}{},
		pingInterval:      interval,
		pingTimeout:       timeout,
		pingMaxWait:       maxWait,
		pingProgressAware: aware,
	}
}

// driveProgress advances clock to each ping and check deadline until `until`
// or until pingLoopProgress returns.
//
// Tickers are created together at start, so deadlines stay aligned when each
// travel lands exactly on the next one. BytesRead is sampled once while the
// tracker is constructed and once per ticker event.
//
// expectDead waits for the disconnect once virtual time is past
// PingInterval+PingTimeout, before another travel can move the clock.
func driveProgress(
	t *testing.T,
	clock *neo.Time,
	fc *fakeTransport,
	errc <-chan error,
	interval, check, timeout, until time.Duration,
	expectDead bool,
	beforeTravel func(target time.Time),
) (time.Duration, error) {
	t.Helper()

	if err := pollSamples(t, fc, 1, errc); err != nil { // newPingTracker calls progress()
		return 0, err
	}
	start := clock.Now()
	nextSend := start.Add(interval)
	nextCheck := start.Add(check)
	samples := fc.samples.Load()
	end := start.Add(until)

	for clock.Now().Before(end) {
		target := nextSend
		if nextCheck.Before(target) {
			target = nextCheck
		}
		if target.After(end) {
			break
		}
		n := 0
		if !nextSend.After(target) {
			n++
			nextSend = nextSend.Add(interval)
		}
		if !nextCheck.After(target) {
			n++
			nextCheck = nextCheck.Add(check)
		}
		if beforeTravel != nil {
			beforeTravel(target)
		}
		samples += uint64(n)
		clock.Travel(target.Sub(clock.Now()))
		if err := pollSamples(t, fc, samples, errc); err != nil {
			return clock.Now().Sub(start), err
		}
		// The handler samples progress before it decides. Once the no-progress
		// deadline can have been crossed, block until that decision is visible
		// so the next travel cannot race with loop exit.
		if expectDead && clock.Now().Sub(start.Add(interval)) > timeout {
			return clock.Now().Sub(start), waitErr(t, errc)
		}
		select {
		case err := <-errc:
			return clock.Now().Sub(start), err
		default:
		}
	}
	return clock.Now().Sub(start), nil
}

// pollSamples waits until BytesRead has been called at least n times.
// If the ping loop returns first, that error is returned instead.
func pollSamples(t *testing.T, fc *fakeTransport, n uint64, errc <-chan error) error {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for fc.samples.Load() < n {
		select {
		case err := <-errc:
			return err
		default:
		}
		if time.Now().After(deadline) {
			select {
			case err := <-errc:
				return err
			default:
				t.Fatalf("timed out waiting for %d progress samples, have %d", n, fc.samples.Load())
			}
		}
		time.Sleep(time.Millisecond)
	}
	return nil
}

func waitErr(t *testing.T, errc <-chan error) error {
	t.Helper()
	select {
	case err := <-errc:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("ping loop did not return")
		return nil
	}
}

func TestPingLoopProgressSlowLink(t *testing.T) {
	const (
		interval = time.Minute
		timeout  = 15 * time.Second
		until    = 5 * time.Minute
	)
	check := pingCheckInterval(timeout)
	clock := neo.NewTime(time.Date(2026, 10, 28, 17, 0, 0, 0, time.UTC))
	fc := &fakeTransport{}
	conn := newPingTestConn(t, clock, fc, interval, timeout, 0, true, log.Nop)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- conn.pingLoopProgress(ctx) }()

	var nextByte time.Time
	elapsed, err := driveProgress(t, clock, fc, errc, interval, check, timeout, until, false, func(target time.Time) {
		if nextByte.IsZero() {
			nextByte = clock.Now().Add(2 * time.Second)
		}
		for !nextByte.After(target) {
			fc.bytes.Add(1)
			nextByte = nextByte.Add(2 * time.Second)
		}
	})
	require.NoError(t, err)
	require.Equal(t, until, elapsed)

	cancel()
	require.ErrorIs(t, waitErr(t, errc), context.Canceled)

	delay := int((interval + timeout).Seconds())
	sends := fc.sent()
	require.Len(t, sends, int(until/interval))
	for _, req := range sends {
		require.Equal(t, delay, req.DisconnectDelay)
		require.NotZero(t, req.PingID)
	}
	require.NoError(t, fc.decodeErr)
}

func TestPingLoopProgressDeadLink(t *testing.T) {
	const (
		interval = time.Minute
		timeout  = 15 * time.Second
	)
	check := pingCheckInterval(timeout)
	clock := neo.NewTime(time.Date(2026, 10, 28, 17, 0, 0, 0, time.UTC))
	logs := &captureLog{}
	fc := &fakeTransport{}
	conn := newPingTestConn(t, clock, fc, interval, timeout, 0, true, logs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- conn.pingLoopProgress(ctx) }()

	elapsed, err := driveProgress(t, clock, fc, errc, interval, check, timeout, interval+timeout+check, true, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "disconnect (pong missed)")
	require.Contains(t, err.Error(), "no_progress")
	require.Greater(t, elapsed, interval+timeout)
	require.LessOrEqual(t, elapsed, interval+timeout+check)
	require.True(t, logs.has("Ping failed, connection considered dead"))
}

func TestPingLoopProgressPong(t *testing.T) {
	const (
		interval = time.Minute
		timeout  = 15 * time.Second
		until    = 5 * time.Minute
	)
	check := pingCheckInterval(timeout)
	clock := neo.NewTime(time.Date(2026, 10, 28, 17, 0, 0, 0, time.UTC))
	fc := &fakeTransport{}
	conn := newPingTestConn(t, clock, fc, interval, timeout, 0, true, log.Nop)
	fc.onSend = func(req mt.PingDelayDisconnectRequest) {
		var b bin.Buffer
		require.NoError(t, b.Encode(&mt.Pong{MsgID: 1, PingID: req.PingID}))
		require.NoError(t, conn.handlePong(&b))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- conn.pingLoopProgress(ctx) }()

	_, err := driveProgress(t, clock, fc, errc, interval, check, timeout, until, false, nil)
	require.NoError(t, err)

	cancel()
	require.ErrorIs(t, waitErr(t, errc), context.Canceled)
	require.NotEmpty(t, fc.sent())
	require.NoError(t, fc.decodeErr)
}

func TestPingLoopProgressWriteError(t *testing.T) {
	const (
		interval = time.Minute
		timeout  = 15 * time.Second
	)
	clock := neo.NewTime(time.Date(2026, 10, 28, 17, 0, 0, 0, time.UTC))
	logs := &captureLog{}
	fc := &fakeTransport{sendErr: io.ErrClosedPipe}
	conn := newPingTestConn(t, clock, fc, interval, timeout, 0, true, logs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- conn.pingLoopProgress(ctx) }()

	// One jump: the send ticker and the pending check ticker each fire once.
	if err := pollSamples(t, fc, 1, errc); err != nil {
		t.Fatalf("loop exited before the first ping: %v", err)
	}
	clock.Travel(interval)
	err := waitErr(t, errc)
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Contains(t, err.Error(), "disconnect (ping write)")
	require.True(t, logs.has("Ping failed, connection considered dead"))
}

func TestPingLoopLegacyMissedPongDespiteProgress(t *testing.T) {
	const (
		interval = time.Minute
		timeout  = 50 * time.Millisecond
	)
	clock := neo.NewTime(time.Date(2026, 10, 28, 17, 0, 0, 0, time.UTC))
	logs := &captureLog{}
	fc := &fakeTransport{}
	conn := newPingTestConn(t, clock, fc, interval, timeout, 0, false, logs)

	sent := make(chan struct{})
	fc.onSend = func(mt.PingDelayDisconnectRequest) {
		fc.bytes.Add(1000)
		close(sent)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	// Ticker creation closes ready only after neo has published the ticker id.
	// Traveling on Observe races with that publication.
	ready := make(chan struct{})
	conn.clock = &readyClock{Time: clock, ready: ready}
	go func() { errc <- conn.pingLoop(ctx) }()

	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("ping ticker was not scheduled")
	}
	clock.Travel(interval)

	select {
	case <-sent:
	case <-time.After(2 * time.Second):
		t.Fatal("ping was not written")
	}
	// More inbound bytes while the pong is outstanding. Legacy mode must
	// ignore them and still declare the connection dead.
	fc.bytes.Add(5000)

	select {
	case err := <-errc:
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Contains(t, err.Error(), "disconnect (pong missed)")
	case <-time.After(2 * time.Second):
		t.Fatal("legacy ping loop did not time out")
	}
	require.True(t, logs.has("Ping failed, connection considered dead"))
	require.Greater(t, fc.bytes.Load(), uint64(0))
}

func TestConnProgress(t *testing.T) {
	plain := &constantConn{}
	c := &Conn{conn: plain}
	require.Equal(t, uint64(0), c.progress())
	c.recvMessages.Inc()
	require.Equal(t, uint64(1), c.progress())

	fc := &fakeTransport{}
	fc.bytes.Store(10)
	c.conn = fc
	require.Equal(t, uint64(11), c.progress())
}
