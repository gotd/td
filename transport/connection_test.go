package transport

import (
	"bytes"
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto/codec"
)

func TestConnection(t *testing.T) {
	leftConn, rightConn := net.Pipe()
	intermediate := codec.Intermediate{}

	left := newConnection(leftConn, intermediate)
	right := newConnection(rightConn, intermediate)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	buf := bytes.Repeat([]byte{1, 2, 3, 4}, 50)
	done := make(chan []byte)
	go func() {
		defer close(done)

		var b bin.Buffer
		if err := right.Recv(ctx, &b); err != nil {
			t.Error(err)
			return
		}

		done <- b.Buf
	}()

	if err := left.Send(ctx, &bin.Buffer{Buf: buf}); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, buf, <-done)
}

// stubConn is a net.Conn whose Write parks until released or closed.
type stubConn struct {
	mux    sync.Mutex
	ops    []string
	writes int

	readDeadline  time.Time
	writeDeadline time.Time

	closedCh  chan struct{}
	closeOnce sync.Once

	writeEntered chan struct{}
	writeRelease chan struct{}
}

func newStubConn() *stubConn {
	return &stubConn{
		closedCh:     make(chan struct{}),
		writeEntered: make(chan struct{}, 1),
		writeRelease: make(chan struct{}),
	}
}

func (c *stubConn) record(op string) {
	c.mux.Lock()
	c.ops = append(c.ops, op)
	c.mux.Unlock()
}

func (c *stubConn) snapshot() []string {
	c.mux.Lock()
	defer c.mux.Unlock()
	return append([]string(nil), c.ops...)
}

func (c *stubConn) writeCount() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.writes
}

func (c *stubConn) Read(b []byte) (int, error) {
	<-c.closedCh
	return 0, net.ErrClosed
}

func (c *stubConn) Write(b []byte) (int, error) {
	c.mux.Lock()
	c.writes++
	c.mux.Unlock()

	select {
	case c.writeEntered <- struct{}{}:
	default:
	}

	select {
	case <-c.writeRelease:
		return len(b), nil
	case <-c.closedCh:
		return 0, net.ErrClosed
	}
}

func (c *stubConn) Close() error {
	c.record("close")
	c.closeOnce.Do(func() { close(c.closedCh) })
	return nil
}

func (c *stubConn) LocalAddr() net.Addr  { return nil }
func (c *stubConn) RemoteAddr() net.Addr { return nil }

func (c *stubConn) SetDeadline(t time.Time) error {
	c.record("deadline")
	return nil
}

func (c *stubConn) SetReadDeadline(t time.Time) error {
	c.mux.Lock()
	c.ops = append(c.ops, "read-deadline")
	c.readDeadline = t
	c.mux.Unlock()
	return nil
}

func (c *stubConn) SetWriteDeadline(t time.Time) error {
	c.mux.Lock()
	c.ops = append(c.ops, "write-deadline")
	c.writeDeadline = t
	c.mux.Unlock()
	return nil
}

func (c *stubConn) readDeadlineValue() time.Time {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.readDeadline
}

func (c *stubConn) writeDeadlineValue() time.Time {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.writeDeadline
}

// TestSendDoesNotBlockOnCancelledContext verifies that a parked Send does not
// trap other senders: the second Send must observe ctx cancellation while
// waiting for the write lock and must never reach the codec.
func TestSendDoesNotBlockOnCancelledContext(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})

	parked := make(chan error, 1)
	go func() {
		// Buf must be a multiple of 4 bytes: the Intermediate codec's
		// checkAlign rejects anything else before it ever reaches Write,
		// which would make this Send never park.
		b := &bin.Buffer{Buf: []byte("frst")}
		parked <- conn.Send(context.Background(), b)
	}()

	// Wait until the first Send is actually inside stub.Write.
	select {
	case <-stub.writeEntered:
	case <-time.After(time.Second):
		t.Fatal("first Send did not reach Write")
	}
	writesAfterPark := stub.writeCount()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		b := &bin.Buffer{Buf: []byte("scnd")}
		done <- conn.Send(ctx, b)
	}()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("second Send blocked on write lock despite cancelled context")
	}

	require.Equal(t, writesAfterPark, stub.writeCount(),
		"second Send must not reach the underlying conn")

	require.NoError(t, conn.Close())
	<-parked
}

// TestSendObservesCancellationWhileParkedOnSemaphore verifies that the
// second Send's ctx cancellation is observed by the select on c.writeSem/
// ctx.Done() itself, not by the earlier ctx.Err() pre-check. The ctx is
// cancelled only after the second Send has already parked waiting for the
// semaphore, so it must be the select that reacts. Without this test, a
// regression that replaced both select acquisitions with a plain blocking
// `c.writeSem <- struct{}{}` — semantically the original uncancellable
// mutex — would leave the whole suite green: only the pre-check path was
// covered before.
func TestSendObservesCancellationWhileParkedOnSemaphore(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})

	parked := make(chan error, 1)
	go func() {
		// Buf must be a multiple of 4 bytes: the Intermediate codec's
		// checkAlign rejects anything else before it ever reaches Write,
		// which would make this Send never park.
		b := &bin.Buffer{Buf: []byte("frst")}
		parked <- conn.Send(context.Background(), b)
	}()

	// Wait until the first Send is actually inside stub.Write, holding
	// writeSem.
	select {
	case <-stub.writeEntered:
	case <-time.After(time.Second):
		t.Fatal("first Send did not reach Write")
	}
	writesAfterPark := stub.writeCount()

	// A still-live ctx: the second Send must pass the ctx.Err() pre-check
	// and reach the select, where it parks because writeSem is held.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		b := &bin.Buffer{Buf: []byte("scnd")}
		done <- conn.Send(ctx, b)
	}()

	// Confirm the second Send is genuinely parked (has not returned) before
	// cancelling: it can only unblock via the semaphore (held by the first
	// Send) or ctx.Done() (not yet fired), so a correct implementation
	// cannot return here.
	select {
	case err := <-done:
		t.Fatalf("second Send returned before ctx was cancelled: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("second Send did not observe ctx cancellation while parked on write lock")
	}

	require.Equal(t, writesAfterPark, stub.writeCount(),
		"second Send must not reach the underlying conn")

	require.NoError(t, conn.Close())
	<-parked
}

// TestSendPreCheckRejectsCancelledContextWithFreeSemaphore verifies that the
// ctx.Err() pre-check at the top of Send is what rejects an already-cancelled
// ctx, not a lucky outcome of the select that follows it. With a free
// semaphore and an already-cancelled ctx, a select with both cases ready
// (c.writeSem <- struct{}{} and <-ctx.Done()) picks pseudo-randomly, so
// removing the pre-check would let roughly half the calls reach the wire.
// 200 iterations make that non-determinism visible reliably instead of
// passing by chance.
func TestSendPreCheckRejectsCancelledContextWithFreeSemaphore(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for i := range 200 {
		err := conn.Send(ctx, &bin.Buffer{Buf: []byte("data")})
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 0, stub.writeCount(),
			"iteration %d: Send must never reach the underlying conn with an already-cancelled ctx", i)
	}
}

// TestCloseForcesDeadlinesBeforeClosing verifies that Close pushes both
// deadlines into the past BEFORE closing, so a pending Write is unblocked even
// on a net.Conn implementation that does not unblock on Close by itself.
func TestCloseForcesDeadlinesBeforeClosing(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})

	before := time.Now()
	require.NoError(t, conn.Close())

	ops := stub.snapshot()
	require.Contains(t, ops, "write-deadline")
	require.Contains(t, ops, "read-deadline")

	closeIdx := -1
	writeDeadlineIdx := -1
	for i, op := range ops {
		switch op {
		case "close":
			if closeIdx == -1 {
				closeIdx = i
			}
		case "write-deadline":
			if writeDeadlineIdx == -1 {
				writeDeadlineIdx = i
			}
		}
	}
	require.NotEqual(t, -1, closeIdx)
	require.NotEqual(t, -1, writeDeadlineIdx)
	require.Less(t, writeDeadlineIdx, closeIdx,
		"deadline must be forced before Close")

	// The headline behavior: the deadlines Close installs must actually be in
	// the past, not merely "set to some value" or cleared (time.Time{} counts
	// as "no deadline", not "in the past", and IsZero() catches that).
	readDeadline := stub.readDeadlineValue()
	writeDeadline := stub.writeDeadlineValue()
	require.False(t, readDeadline.IsZero(), "read deadline must be set, not cleared")
	require.False(t, writeDeadline.IsZero(), "write deadline must be set, not cleared")
	require.True(t, readDeadline.Before(before), "read deadline must be in the past")
	require.True(t, writeDeadline.Before(before), "write deadline must be in the past")
}

// signalingConn wraps a net.Conn and signals writeEntered right before
// forwarding to the underlying Write, so a test can deterministically wait
// until a goroutine has reached Write instead of guessing with a sleep.
type signalingConn struct {
	net.Conn
	writeEntered chan struct{}
}

func (c *signalingConn) Write(b []byte) (int, error) {
	select {
	case c.writeEntered <- struct{}{}:
	default:
	}
	return c.Conn.Write(b)
}

// TestCloseUnblocksParkedSendOnRealConn is the end-to-end version on a real
// net.Pipe pair: nobody is reading, so Write parks; Close must release it.
func TestCloseUnblocksParkedSendOnRealConn(t *testing.T) {
	left, right := net.Pipe()
	defer func() { _ = right.Close() }()

	wrapped := &signalingConn{Conn: left, writeEntered: make(chan struct{}, 1)}
	conn := newConnection(wrapped, codec.Intermediate{})

	done := make(chan error, 1)
	go func() {
		b := &bin.Buffer{Buf: make([]byte, 1024)}
		done <- conn.Send(context.Background(), b)
	}()

	// Wait until the Send is actually inside Write instead of guessing with
	// a sleep: on a loaded machine a fixed sleep may fire before the
	// goroutine parks, making the test pass for the wrong reason.
	select {
	case <-wrapped.writeEntered:
	case <-time.After(time.Second):
		t.Fatal("Send did not reach Write")
	}

	require.NoError(t, conn.Close())

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not unblock parked Send")
	}
}

// deadlineStubConn is a net.Conn whose Close is a no-op — it does not unblock
// pending I/O by itself, modeling wrapper net.Conn implementations
// (websocket, mtproxy obfuscator) that gotd layers transport over. Its Write
// unblocks only once a write deadline in the past has actually been set, so
// this is the true regression test for the deadline-forcing behavior: it can
// only pass because Close pushed the deadline into the past, never because
// Close itself released the connection.
type deadlineStubConn struct {
	writeEntered chan struct{}

	pastDeadlineCh   chan struct{}
	pastDeadlineOnce sync.Once

	closedCh  chan struct{}
	closeOnce sync.Once
}

func newDeadlineStubConn() *deadlineStubConn {
	return &deadlineStubConn{
		writeEntered:   make(chan struct{}, 1),
		pastDeadlineCh: make(chan struct{}),
		closedCh:       make(chan struct{}),
	}
}

func (c *deadlineStubConn) Read(b []byte) (int, error) {
	<-c.closedCh
	return 0, net.ErrClosed
}

func (c *deadlineStubConn) Write(b []byte) (int, error) {
	select {
	case c.writeEntered <- struct{}{}:
	default:
	}

	<-c.pastDeadlineCh
	return 0, os.ErrDeadlineExceeded
}

func (c *deadlineStubConn) Close() error {
	c.closeOnce.Do(func() { close(c.closedCh) })
	return nil
}

func (c *deadlineStubConn) LocalAddr() net.Addr  { return nil }
func (c *deadlineStubConn) RemoteAddr() net.Addr { return nil }

func (c *deadlineStubConn) SetDeadline(t time.Time) error { return nil }

func (c *deadlineStubConn) SetReadDeadline(t time.Time) error { return nil }

func (c *deadlineStubConn) SetWriteDeadline(t time.Time) error {
	if !t.IsZero() && t.Before(time.Now()) {
		c.pastDeadlineOnce.Do(func() { close(c.pastDeadlineCh) })
	}
	return nil
}

// TestSendAfterCloseDoesNotClearForcedDeadline is the regression test for the
// deadline-erasure race: without the closed check running under the same
// mutex as Close's deadline-forcing, a Send that starts after Close would
// still call SetWriteDeadline(time.Time{}) and erase the forced past
// deadline before observing "closed", parking a subsequent write forever on
// a net.Conn that does not unblock on Close by itself.
func TestSendAfterCloseDoesNotClearForcedDeadline(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})
	require.NoError(t, conn.Close())
	_ = conn.Send(context.Background(), &bin.Buffer{Buf: make([]byte, 4)})
	require.False(t, stub.writeDeadlineValue().IsZero(),
		"Send after Close must not clear the forced past deadline")
}

// TestRecvAfterCloseDoesNotClearForcedDeadline is the Recv counterpart of
// TestSendAfterCloseDoesNotClearForcedDeadline.
func TestRecvAfterCloseDoesNotClearForcedDeadline(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})
	require.NoError(t, conn.Close())
	var b bin.Buffer
	_ = conn.Recv(context.Background(), &b)
	require.False(t, stub.readDeadlineValue().IsZero(),
		"Recv after Close must not clear the forced past deadline")
}

// TestSendAfterCloseIsRetryable verifies that a Send failing because the
// connection was already closed is classified as ErrWriteFailed: the message
// was never handed to codec.Write, so retrying it on a new connection is
// safe. net.ErrClosed must remain observable in the chain too.
func TestSendAfterCloseIsRetryable(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})
	require.NoError(t, conn.Close())

	err := conn.Send(context.Background(), &bin.Buffer{Buf: make([]byte, 4)})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrWriteFailed)
	require.ErrorIs(t, err, net.ErrClosed)
}

// TestWriteErrorIsSingleLine pins the rendering of write failures: they must
// occupy exactly one log line. errors.Join, used here previously, separates
// joined errors with a newline, which breaks grep and line-oriented log
// ingestion. Both classifications must survive the change.
func TestWriteErrorIsSingleLine(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})
	require.NoError(t, conn.Close())

	err := conn.Send(context.Background(), &bin.Buffer{Buf: make([]byte, 4)})
	require.Error(t, err)
	require.NotContains(t, err.Error(), "\n", "write error must render on a single line")
	require.Contains(t, err.Error(), ErrWriteFailed.Error())
	require.Contains(t, err.Error(), net.ErrClosed.Error())
	require.ErrorIs(t, err, ErrWriteFailed)
	require.ErrorIs(t, err, net.ErrClosed)
}

// TestRecvAfterCloseIsNotWriteFailed verifies the opposite for Recv: a
// failed read must never be classified as ErrWriteFailed, since it is not a
// write and callers must not treat it as "safe to retry the send".
func TestRecvAfterCloseIsNotWriteFailed(t *testing.T) {
	stub := newStubConn()
	conn := newConnection(stub, codec.Intermediate{})
	require.NoError(t, conn.Close())

	var b bin.Buffer
	err := conn.Recv(context.Background(), &b)
	require.Error(t, err)
	require.ErrorIs(t, err, net.ErrClosed)
	require.NotErrorIs(t, err, ErrWriteFailed)
}

// TestSendCodecRejectionIsNotWriteFailed pins the boundary of ErrWriteFailed.
//
// A message the codec refuses is a property of the message, not of the link:
// the rejection happens before the socket is touched, so the connection stays
// healthy. Marking it as a transport failure would tell callers
// (telegram/invoke.go, pool/pool_conn.go) that the frame never reached the
// server and is safe to resend on a new connection — but the resend fails
// identically, so the caller would park waiting for a reconnect that nothing
// triggers, and a pool would spin marking connections dead instead.
func TestSendCodecRejectionIsNotWriteFailed(t *testing.T) {
	for _, tt := range []struct {
		name string
		buf  *bin.Buffer
		want error
	}{
		{"empty message", &bin.Buffer{}, codec.ErrInvalidMessageLength},
		{"oversized message", &bin.Buffer{Buf: make([]byte, (1<<24)+4)}, codec.ErrInvalidMessageLength},
		{"misaligned payload", &bin.Buffer{Buf: make([]byte, 3)}, codec.ErrPayloadNotAligned},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stub := newStubConn()
			conn := newConnection(stub, codec.Intermediate{})

			err := conn.Send(context.Background(), tt.buf)
			require.Error(t, err)
			require.ErrorIs(t, err, tt.want)
			require.NotErrorIs(t, err, ErrWriteFailed,
				"a locally rejected message must not be reported as a transport failure")
			require.Equal(t, 0, stub.writeCount(),
				"the codec rejected the message, so nothing may reach the socket")
		})
	}
}

// TestCloseUnblocksOnlyViaForcedDeadline proves that it is specifically the
// forced past deadline — not Close itself — that releases a parked Send. The
// stub's Close never unblocks pending I/O on its own, so if Close stopped
// forcing a past write deadline, this test would hang until its own timeout
// and fail.
func TestCloseUnblocksOnlyViaForcedDeadline(t *testing.T) {
	stub := newDeadlineStubConn()
	conn := newConnection(stub, codec.Intermediate{})

	done := make(chan error, 1)
	go func() {
		b := &bin.Buffer{Buf: make([]byte, 4)}
		done <- conn.Send(context.Background(), b)
	}()

	select {
	case <-stub.writeEntered:
	case <-time.After(time.Second):
		t.Fatal("Send did not reach Write")
	}

	require.NoError(t, conn.Close())

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not unblock parked Send via forced deadline")
	}
}
