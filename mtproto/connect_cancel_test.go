package mtproto

import (
	"context"
	"crypto/rand"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/transport"
)

// hangingConn never returns from Recv until closed. It models a raw socket
// read that ignores ctx entirely and can only be interrupted by Close().
type hangingConn struct {
	closed    chan struct{}
	started   chan struct{}
	once      sync.Once
	closeOnce sync.Once
}

func newHangingConn() *hangingConn {
	return &hangingConn{
		closed:  make(chan struct{}),
		started: make(chan struct{}),
	}
}

func (c *hangingConn) Send(ctx context.Context, b *bin.Buffer) error { return nil }

func (c *hangingConn) Recv(ctx context.Context, b *bin.Buffer) error {
	c.once.Do(func() { close(c.started) })
	<-c.closed
	return context.Canceled
}

func (c *hangingConn) Close() error {
	c.closeOnce.Do(func() { close(c.closed) })
	return nil
}

var _ transport.Conn = (*hangingConn)(nil)

// TestConnectIsInterruptibleByContext guards the production incident: while
// connect() runs the key exchange, Conn.Run has not started its goroutine
// group yet, so handleClose does not exist and a cancelled ctx cannot close
// the socket on its own. pool.DC.Close() (cancel(), then Supervisor.Wait)
// therefore hung forever.
func TestConnectIsInterruptibleByContext(t *testing.T) {
	hanging := newHangingConn()

	conn := New(func(ctx context.Context) (transport.Conn, error) {
		return hanging, nil
	}, Options{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- conn.Run(ctx, func(ctx context.Context) error { return nil })
	}()

	// Wait until connect() is actually parked in the exchange read (step 2:
	// readUnencrypted for ResPQ), rather than guessing with a sleep.
	select {
	case <-hanging.started:
	case <-time.After(2 * time.Second):
		t.Fatal("connect() never reached the exchange read")
	}

	cancel()

	select {
	case err := <-done:
		// The watcher's forced Close() breaks the parked read with a raw
		// transport error; connect() must normalize that back to ctx.Err()
		// so callers classifying errors via errors.Is see cancellation, not
		// a transport failure.
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("connect did not return after context cancellation")
	}
}

// TestConnectWatcherDoesNotLeakOnSuccess proves that the ctx-watcher spawned
// by connect() terminates as soon as connect() returns successfully, rather
// than lingering for the lifetime of the process. A goroutine leaked per
// connection attempt would be worse than the hang this watcher fixes.
func TestConnectWatcherDoesNotLeakOnSuccess(t *testing.T) {
	a := require.New(t)

	closeMe := &closeConn{}
	c := Conn{
		dialer: func(ctx context.Context) (transport.Conn, error) {
			return closeMe, nil
		},
		clock: clock.System,
		authKey: crypto.AuthKey{
			ID: [8]byte{1}, // Skip exchange.
		},
		rand: rand.Reader, // newSessionID succeeds, so connect() returns nil.
		log:  log.For(log.Nop),
		// Real Conns always carry a positive DialTimeout (Options defaults it
		// to 35s; see options.go). Leaving this zero would make connectCtx
		// expire instantly on every iteration below, and the resulting
		// watcher force-close races on closeMe with itself and neighboring
		// iterations of the loop.
		dialTimeout: time.Minute,
	}

	before := runtime.NumGoroutine()

	const iterations = 50
	for range iterations {
		a.NoError(c.connect(t.Context()))
	}

	// Poll in this goroutine rather than via require.Eventually: that helper
	// evaluates the condition in a freshly spawned goroutine, which would
	// itself be counted by runtime.NumGoroutine() and make the check
	// unsatisfiable regardless of whether the watcher leaks.
	deadline := time.Now().Add(time.Second)
	var after int
	for {
		after = runtime.NumGoroutine()
		if after <= before || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	a.LessOrEqual(after, before, "watcher goroutines leaked after successful connect()")
}

// TestConnectDialTimeoutDuringExchangeNormalizesToDeadlineExceeded guards the
// path the watcher was widened to cover: connectCtx (dialTimeout), not just
// ctx, can be what force-closes a parked exchange read. Before widening the
// normalization defer to also check connectCtx.Err(), this path surfaced the
// watcher's raw "use of closed network connection"-shaped transport error
// instead of a wrapped context.DeadlineExceeded, breaking callers that
// classify errors via errors.Is.
func TestConnectDialTimeoutDuringExchangeNormalizesToDeadlineExceeded(t *testing.T) {
	a := require.New(t)

	hanging := newHangingConn()

	c := New(func(ctx context.Context) (transport.Conn, error) {
		return hanging, nil
	}, Options{DialTimeout: 200 * time.Millisecond})

	// The caller's ctx is never cancelled; only the internal dial timeout
	// fires while the exchange read is parked in hanging.Recv.
	err := c.connect(context.Background())
	a.Error(err)
	a.ErrorIs(err, context.DeadlineExceeded)
}

// TestConnectGenuineExchangeErrorIsNotRewritten guards the defer's exact
// declaration point (must run AFTER connectCtx's own "defer cancel()" has
// NOT yet fired, i.e. be declared right after that cancel() so it runs
// before it -- see the placement comment in connect.go). A defer declared
// before connectCtx's cancel() would observe connectCtx already cancelled on
// every return path, including this one, and would rewrite this genuine,
// non-context exchange failure into context.Canceled. Both ctx and
// connectCtx stay alive for the whole (synchronous, non-blocking) call, so
// this failure must surface unchanged.
func TestConnectGenuineExchangeErrorIsNotRewritten(t *testing.T) {
	a := require.New(t)

	failing := &closeConn{}
	c := New(func(ctx context.Context) (transport.Conn, error) {
		return failing, nil
	}, Options{DialTimeout: time.Minute})

	err := c.connect(context.Background())
	a.Error(err)
	a.ErrorIs(err, io.EOF, "genuine exchange error must survive unchanged")
	a.NotErrorIs(err, context.Canceled)
	a.NotErrorIs(err, context.DeadlineExceeded)
}
