package mtproto

import (
	"context"
	"crypto/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log"
	"github.com/gotd/log/logzap"
	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

// testLogger builds a log.Helper backed by zaptest, so failures print through
// t.Log. Conn.log is a log.Helper (github.com/gotd/log), not a *zap.Logger.
func testLogger(t *testing.T) log.Helper {
	return log.For(logzap.New(zaptest.NewLogger(t)))
}

func TestIdleWatchdogFiresOnSilence(t *testing.T) {
	clk := neo.NewTime(testutil.Date())
	observer := clk.Observe()
	hanging := newHangingConn()

	c := &Conn{
		conn:        hanging,
		clock:       clk,
		idleTimeout: time.Minute,
		log:         testLogger(t),
	}
	c.lastRecv.Store(clk.Now().UnixNano())

	done := make(chan error, 1)
	go func() {
		done <- c.idleWatchdog(context.Background())
	}()

	// Wait until the watchdog registered its ticker, then jump past the
	// timeout. Without this handshake the test races the ticker registration.
	<-observer
	clk.Travel(2 * time.Minute)

	select {
	case err := <-done:
		require.True(t, errors.Is(err, ErrIdleTimeout), "got %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("watchdog did not fire on silence")
	}

	// Watchdog must close the connection so parked I/O is released.
	select {
	case <-hanging.closed:
	default:
		t.Fatal("watchdog did not close the connection")
	}
}

func TestIdleWatchdogStaysQuietWhileReceiving(t *testing.T) {
	clk := neo.NewTime(testutil.Date())
	observer := clk.Observe()
	hanging := newHangingConn()

	c := &Conn{
		conn:        hanging,
		clock:       clk,
		idleTimeout: time.Minute,
		log:         testLogger(t),
	}
	c.lastRecv.Store(clk.Now().UnixNano())

	done := make(chan error, 1)
	go func() {
		done <- c.idleWatchdog(context.Background())
	}()

	<-observer
	// Advance in steps shorter than the timeout, refreshing lastRecv each time
	// the way readLoop does.
	for range 4 {
		clk.Travel(20 * time.Second)
		c.lastRecv.Store(clk.Now().UnixNano())
	}

	select {
	case err := <-done:
		t.Fatalf("watchdog fired while data was flowing: %v", err)
	case <-time.After(200 * time.Millisecond):
		// Expected: still running.
	}
}

// TestIdleWatchdogTerminatesOnContextCancel guards the goroutine-group
// integration: idleWatchdog runs inside Run's tdsync.LogGroup, so it must
// exit as soon as the group's context is cancelled, the same as every other
// loop in that group (pingLoop, ackLoop, saltLoop, handleClose).
func TestIdleWatchdogTerminatesOnContextCancel(t *testing.T) {
	clk := neo.NewTime(testutil.Date())
	c := &Conn{
		conn:        newHangingConn(),
		clock:       clk,
		idleTimeout: time.Minute,
		log:         testLogger(t),
	}
	c.lastRecv.Store(clk.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- c.idleWatchdog(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("idleWatchdog did not terminate on context cancellation")
	}
}

// TestIdleWatchdogNoGoroutineLeakAcrossCycles guards against idleWatchdog
// itself leaking a goroutine across repeated Run cycles: idleWatchdog is
// created and torn down once per Run call, so a per-call leak would
// accumulate over the lifetime of a long-running process that reconnects
// often. It only observes goroutine count via runtime.NumGoroutine() and so
// cannot detect a leaked clock.Timer -- a Timer leak is a runtime timer, not
// a goroutine, and defer timer.Stop() in idleWatchdog (conn.go) guards
// against that separately. Mirrors the polling pattern in
// TestConnectWatcherDoesNotLeakOnSuccess (connect_cancel_test.go).
func TestIdleWatchdogNoGoroutineLeakAcrossCycles(t *testing.T) {
	a := require.New(t)

	before := runtime.NumGoroutine()

	const iterations = 50
	for range iterations {
		c := &Conn{
			conn:        newHangingConn(),
			clock:       clock.System,
			idleTimeout: time.Minute,
			log:         testLogger(t),
		}
		c.lastRecv.Store(clock.System.Now().UnixNano())

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- c.idleWatchdog(ctx)
		}()
		cancel()
		<-done
	}

	// Poll in this goroutine rather than via require.Eventually: that helper
	// evaluates the condition in a freshly spawned goroutine, which would
	// itself be counted by runtime.NumGoroutine() and make the check
	// unsatisfiable regardless of whether the watchdog leaks.
	deadline := time.Now().Add(time.Second)
	var after int
	for {
		after = runtime.NumGoroutine()
		if after <= before || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	a.LessOrEqual(after, before, "idleWatchdog goroutines leaked across repeated cycles")
}

// stopTrackingTimer wraps a clock.Timer and records whether Stop was called
// on it, so a test can observe idleWatchdog's own Stop() call rather than
// inferring it indirectly (e.g. via goroutine count, which cannot see a
// live but unstopped Timer at all).
type stopTrackingTimer struct {
	clock.Timer
	stopped *atomic.Bool
}

func (t stopTrackingTimer) Stop() bool {
	t.stopped.Store(true)
	return t.Timer.Stop()
}

// stopTrackingClock wraps a clock.Clock, tagging every Timer it creates with
// stopTrackingTimer.
type stopTrackingClock struct {
	clock.Clock
	stopped *atomic.Bool
}

func (c stopTrackingClock) Timer(d time.Duration) clock.Timer {
	return stopTrackingTimer{Timer: c.Clock.Timer(d), stopped: c.stopped}
}

// TestIdleWatchdogStopsTimerOnExit genuinely detects an unstopped Timer,
// unlike TestIdleWatchdogNoGoroutineLeakAcrossCycles above: it wraps the
// Timer idleWatchdog creates and asserts Stop() was actually called on
// context-cancel exit. Deleting `defer timer.Stop()` in idleWatchdog
// (conn.go) fails this test; it does not fail the goroutine-count test,
// since a leaked Timer is a runtime timer, not a goroutine.
func TestIdleWatchdogStopsTimerOnExit(t *testing.T) {
	stopped := &atomic.Bool{}
	c := &Conn{
		conn:        newHangingConn(),
		clock:       stopTrackingClock{Clock: clock.System, stopped: stopped},
		idleTimeout: time.Minute,
		log:         testLogger(t),
	}
	c.lastRecv.Store(clock.System.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- c.idleWatchdog(ctx)
	}()
	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("idleWatchdog did not terminate on context cancellation")
	}

	require.True(t, stopped.Load(), "idleWatchdog must Stop() its Timer on exit")
}

// TestConnectInitializesLastRecv guards against the watchdog firing
// spuriously on a freshly established connection: connect() must set
// lastRecv itself, since a connection that has not read anything yet would
// otherwise look exactly like one that has gone silent.
func TestConnectInitializesLastRecv(t *testing.T) {
	a := require.New(t)

	c := &Conn{
		dialer: func(ctx context.Context) (transport.Conn, error) {
			return &closeConn{}, nil
		},
		clock: clock.System,
		authKey: crypto.AuthKey{
			ID: [8]byte{1}, // Skip exchange.
		},
		rand:        rand.Reader, // newSessionID succeeds, so connect() returns nil.
		log:         log.For(log.Nop),
		dialTimeout: time.Minute,
	}
	a.Zero(c.lastRecv.Load(), "precondition: lastRecv starts unset")

	before := c.clock.Now().UnixNano()
	a.NoError(c.connect(t.Context()))
	after := c.clock.Now().UnixNano()

	got := c.lastRecv.Load()
	a.NotZero(got, "connect() must initialize lastRecv")
	a.GreaterOrEqual(got, before)
	a.LessOrEqual(got, after)
}

// sessionMismatchCipher lets a test drive readLoop's real receive path
// without needing genuine ciphertext or a real AuthKey: every frame decrypts
// to a message whose SessionID never matches Conn's, so consumeMessage
// rejects it as errRejected -- the same as a benign duplicate/replay frame
// -- and readLoop keeps looping. That is all TestReadLoopRefreshesLastRecv
// needs: only the successful Recv (and the c.lastRecv.Store that follows it
// in read.go) is under test, not message acceptance.
type sessionMismatchCipher struct{}

func (sessionMismatchCipher) DecryptFromBuffer(crypto.AuthKey, *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	return &crypto.EncryptedMessageData{SessionID: 0}, nil
}

func (sessionMismatchCipher) Encrypt(crypto.AuthKey, crypto.EncryptedMessageData, *bin.Buffer) error {
	return nil
}

var _ Cipher = sessionMismatchCipher{}

// pacedConn delivers a successful Recv only when the test signals it via
// release, letting a test drive readLoop on an explicit per-frame schedule
// instead of a busy loop.
type pacedConn struct {
	release   chan struct{}
	closed    chan struct{}
	closeOnce sync.Once
}

func newPacedConn() *pacedConn {
	return &pacedConn{
		release: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

func (c *pacedConn) Send(ctx context.Context, b *bin.Buffer) error { return nil }

func (c *pacedConn) Recv(ctx context.Context, b *bin.Buffer) error {
	select {
	case <-c.release:
		return nil
	case <-c.closed:
		return context.Canceled
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *pacedConn) Close() error {
	c.closeOnce.Do(func() { close(c.closed) })
	return nil
}

var _ transport.Conn = (*pacedConn)(nil)

// TestReadLoopRefreshesLastRecv drives the real production wiring --
// readLoop itself (mtproto/read.go), not a test body poking lastRecv
// directly -- to prove the c.lastRecv.Store call there is what idleWatchdog
// actually depends on. TestIdleWatchdogFiresOnSilence and
// TestIdleWatchdogStaysQuietWhileReceiving both simulate the read loop by
// storing lastRecv from the test body, so they would stay green even if
// that store line were deleted from read.go; this test would not.
func TestReadLoopRefreshesLastRecv(t *testing.T) {
	clk := neo.NewTime(testutil.Date())
	observer := clk.Observe()
	pc := newPacedConn()

	c := &Conn{
		conn:         pc,
		clock:        clk,
		cipher:       sessionMismatchCipher{},
		messageIDBuf: proto.NewMessageIDBuf(100),
		sessionID:    1, // Nonzero: sessionMismatchCipher always reports 0, so every delivered frame is a deterministic reject, never a real message.
		idleTimeout:  time.Minute,
		log:          testLogger(t),
	}
	c.lastRecv.Store(clk.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())

	readDone := make(chan error, 1)
	go func() { readDone <- c.readLoop(ctx) }()

	watchdogDone := make(chan error, 1)
	go func() { watchdogDone <- c.idleWatchdog(ctx) }()
	// t.Cleanup, not a bare defer: require.Eventually below calls
	// t.FailNow() on failure, which unwinds this goroutine via
	// runtime.Goexit() before reaching the trailing cancel/join code past
	// the loop. A bare defer would still fire cancel() but not wait for
	// these goroutines to actually exit, letting them log through
	// testLogger's zaptest sink after the test is already marked done,
	// which panics. t.Cleanup runs (and blocks) regardless of how the test
	// function exits.
	t.Cleanup(func() {
		cancel()
		<-readDone
		<-watchdogDone
	})
	<-observer // idleWatchdog's Timer registered.

	// Advance in steps shorter than idleTimeout, delivering one real frame
	// through readLoop before each step -- same schedule as
	// TestIdleWatchdogStaysQuietWhileReceiving, but the refresh now comes
	// from readLoop's own Recv+Store, not the test body.
	for range 4 {
		before := clk.Now().UnixNano()
		select {
		case pc.release <- struct{}{}:
		case <-time.After(2 * time.Second):
			t.Fatal("readLoop did not accept the delivered frame in time")
		}
		// Wait for readLoop to actually record the frame before advancing
		// the clock again, so the watchdog never observes a stale
		// lastRecv racing the next Travel().
		require.Eventually(t, func() bool {
			return c.lastRecv.Load() >= before
		}, 2*time.Second, time.Millisecond, "readLoop did not refresh lastRecv for the delivered frame")

		clk.Travel(20 * time.Second)
	}

	select {
	case err := <-watchdogDone:
		t.Fatalf("watchdog fired while readLoop kept receiving frames: %v", err)
	case <-time.After(200 * time.Millisecond):
		// Still running: expected. t.Cleanup above tears both goroutines
		// down.
	}
}

// countingCloseConn counts every call to Close(), without deduplicating them
// itself -- unlike hangingConn's internal closeOnce, so that if Conn's own
// shared close guard (Conn.close, mtproto/conn.go) ever regressed back to
// letting handleClose and idleWatchdog each reach the underlying transport's
// Close() independently, a test using this type would observe more than one
// call. Recv blocks until the first Close(), modeling a silent connection
// whose only exit is being force-closed.
type countingCloseConn struct {
	closes    atomic.Int32
	closed    chan struct{}
	closeOnce sync.Once
}

func newCountingCloseConn() *countingCloseConn {
	return &countingCloseConn{closed: make(chan struct{})}
}

func (c *countingCloseConn) Send(ctx context.Context, b *bin.Buffer) error { return nil }

func (c *countingCloseConn) Recv(ctx context.Context, b *bin.Buffer) error {
	<-c.closed
	return context.Canceled
}

func (c *countingCloseConn) Close() error {
	c.closes.Add(1)
	// The underlying channel must only be closed once regardless of how
	// many times Close() itself is called -- this models the idempotent
	// close(2) syscall a real net.Conn wraps, while still counting every
	// call this test's production code makes into it, which is what is
	// actually under test here.
	c.closeOnce.Do(func() { close(c.closed) })
	return nil
}

var _ transport.Conn = (*countingCloseConn)(nil)

// runSilentIdleTimeout starts a full Conn.Run cycle over a transport that
// never sends anything and waits for it to return.
//
// This deliberately uses the real clock.System, not a neo.Time simulated
// clock, and small real durations instead of a big Travel() jump: pingLoop
// and ackLoop (unlike idleWatchdog) still use c.clock.Ticker, and neo's
// Ticker has a genuine data race between a Ticker's construction (in the
// loop's own goroutine) and a concurrent Travel() call (from this test's
// goroutine) reading back its id -- the same upstream hazard idleWatchdog
// was rewritten around (see the comment on idleWatchdog in conn.go), but
// unreachable here since pingLoop/ackLoop are out of scope to change. Real
// time sidesteps it entirely: clock.System's Ticker is a stdlib
// *time.Ticker, safe for concurrent use by design.
//
// PingTimeout is set far beyond this test's own real-time budget so that the
// single ping round trip just blocks harmlessly against the silent conn
// instead of independently erroring pingLoop first.
//
// idleTimeout is written directly rather than passed through Options:
// setDefaults enforces IdleTimeout >= PingInterval+PingTimeout, since the
// watchdog is a backstop that must not preempt pingLoop's own pong timeout on
// a healthy connection. Honoring that floor here would push the watchdog past
// the test's real-time budget. The wiring under test — Run -> idleWatchdog ->
// handleClose — does not care how the timeout was configured, so the
// assertions below are unaffected.
func runSilentIdleTimeout(t *testing.T, conn transport.Conn) error {
	t.Helper()

	c := New(func(ctx context.Context) (transport.Conn, error) {
		return conn, nil
	}, Options{
		PingInterval: 100 * time.Millisecond,
		PingTimeout:  10 * time.Second,
		DialTimeout:  5 * time.Second,
	})
	c.idleTimeout = 300 * time.Millisecond
	// Preset a non-zero auth key and readable rand so connect() skips the
	// real key exchange and Run proceeds straight to its goroutine group --
	// same setup TestConnectWatcherDoesNotLeakOnSuccess uses.
	c.authKey = crypto.AuthKey{ID: [8]byte{1}}
	c.rand = rand.Reader

	done := make(chan error, 1)
	go func() {
		done <- c.Run(context.Background(), func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return while the transport was silent")
		return nil // unreachable
	}
}

// TestRunReturnsErrIdleTimeoutOnSilence exercises the real wiring through
// Conn.Run -- not idleWatchdog called directly -- proving idleWatchdog is
// actually registered in Run's goroutine group
// (g.Go("idleWatchdog", c.idleWatchdog) in mtproto/conn.go). It also
// exercises the fix for the nondeterminism the reviewer measured
// (~1-in-40 runs surfacing a raw transport error instead of
// ErrIdleTimeout): handleClose's forced Close() and idleWatchdog's own
// Close() race to close the same silent connection, and Run now merges
// idleWatchdog's recorded cause (Conn.idleCause) into whichever error
// errgroup happened to record first (see Conn.groupErr), so this
// classification no longer depends on which one won that race.
func TestRunReturnsErrIdleTimeoutOnSilence(t *testing.T) {
	err := runSilentIdleTimeout(t, newHangingConn())
	require.True(t, errors.Is(err, ErrIdleTimeout), "got %v", err)
}

// TestGroupErrMergesIdleCauseWithGroupError proves Finding 1's fix: when
// idleCause is set, groupErr must merge it with errgroup's own error rather
// than discarding that error, so both remain independently discoverable via
// errors.Is on the single returned error.
//
// This matters beyond determinism: telegram/pfs.go:15 uses
// errors.Is(err, mtproto.ErrPFSDropKeysRequired) on Run's returned error to
// decide whether to wipe a stored key that the server has forgotten. Before
// this fix, groupErr replaced errgroup's error outright whenever idleCause
// was set, so a ErrPFSDropKeysRequired recorded by errgroup at the same time
// idleWatchdog fired would have been silently shadowed by ErrIdleTimeout,
// leaving handlePrimaryConnDead unable to see it and the stale key in place.
//
// idleCause is produced by a real idleWatchdog fire (same simulated-clock
// technique as TestIdleWatchdogFiresOnSilence) rather than hand-constructed,
// so this exercises Conn.groupErr -- the same method Run calls -- against a
// genuine production value. The "group's own error" is a distinct sentinel
// supplied directly to groupErr: reproducing the real errgroup race
// deterministically would require winning it against idleWatchdog's own
// Close()-triggered readLoop error, which is exactly the nondeterminism
// TestRunReturnsErrIdleTimeoutOnSilence exists to cover, not something a
// second test should re-derive.
func TestGroupErrMergesIdleCauseWithGroupError(t *testing.T) {
	sentinel := errors.New("distinct sentinel error from errgroup")

	clk := neo.NewTime(testutil.Date())
	observer := clk.Observe()
	hanging := newHangingConn()

	c := &Conn{
		conn:        hanging,
		clock:       clk,
		idleTimeout: time.Minute,
		log:         testLogger(t),
	}
	c.lastRecv.Store(clk.Now().UnixNano())

	done := make(chan error, 1)
	go func() {
		done <- c.idleWatchdog(context.Background())
	}()
	<-observer
	clk.Travel(2 * time.Minute)

	select {
	case err := <-done:
		require.True(t, errors.Is(err, ErrIdleTimeout), "got %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("watchdog did not fire on silence")
	}
	require.NotNil(t, c.idleCause.Load(), "precondition: idleWatchdog must have recorded idleCause")

	got := c.groupErr(sentinel)
	require.True(t, errors.Is(got, ErrIdleTimeout), "idle timeout classification must survive the merge: got %v", got)
	require.True(t, errors.Is(got, sentinel), "the group's own error must survive the merge: got %v", got)
}

// TestRunClosesUnderlyingConnExactlyOnce guards the other half of Finding 3:
// once idleWatchdog fires it closes the connection directly, and then
// handleClose closes it again once the group cancels. transport.Conn does
// not promise concurrent-Close safety, so both must route through the
// shared Conn.close guard -- this asserts the underlying transport's
// Close() is reached exactly once per Run cycle, not twice.
func TestRunClosesUnderlyingConnExactlyOnce(t *testing.T) {
	conn := newCountingCloseConn()

	err := runSilentIdleTimeout(t, conn)
	require.True(t, errors.Is(err, ErrIdleTimeout), "got %v", err)

	require.EqualValues(t, 1, conn.closes.Load(),
		"underlying Close() must be reached exactly once per Run cycle")
}
