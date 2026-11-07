package telegram

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// hangConn never answers requests.
type hangConn struct {
	calls atomic.Int64
}

func (h *hangConn) Run(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }

func (h *hangConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	h.calls.Add(1)
	<-ctx.Done()
	return ctx.Err()
}

func (h *hangConn) Ping(context.Context) error { return nil }

// fakeReconnect emulates reconnection loop: on restart, replaces primary
// connection with next and signals ready.
func fakeReconnect(t *testing.T, c *Client, next clientConn) *atomic.Int64 {
	t.Helper()
	var restarts atomic.Int64
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	go func() {
		for {
			select {
			case <-done:
				return
			case <-c.restart:
				restarts.Add(1)
				c.connMux.Lock()
				c.replaceConn(next)
				c.connMux.Unlock()
				c.onReady()
			}
		}
	}()
	return &restarts
}

func TestClientRequestTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var restarts *atomic.Int64
	newClient := func() (*Client, *hangConn) {
		good := newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
			return &tg.NearestDC{ThisDC: 2}, nil
		})
		c := newTestClient(nil)
		hang := &hangConn{}
		c.conn = hang
		c.ready.Signal()
		c.requestTimeout = func(input bin.Encoder) time.Duration {
			if _, ok := input.(*tg.HelpGetNearestDCRequest); ok {
				return 50 * time.Millisecond
			}
			return 0
		}
		restarts = fakeReconnect(t, c, good.conn)
		return c, hang
	}

	t.Run("Resend", func(t *testing.T) {
		c, hang := newClient()
		r, err := c.API().HelpGetNearestDC(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, r.ThisDC)
		require.EqualValues(t, 1, hang.calls.Load())
	})
	t.Run("Concurrent", func(t *testing.T) {
		c, hang := newClient()
		errs := make(chan error, 3)
		for range 3 {
			go func() {
				_, err := c.API().HelpGetNearestDC(ctx)
				errs <- err
			}()
		}
		for range 3 {
			require.NoError(t, <-errs)
		}
		require.EqualValues(t, 3, hang.calls.Load())
		require.EqualValues(t, 1, restarts.Load(), "one reconnect for all timed out requests")
	})
	t.Run("NoTimeout", func(t *testing.T) {
		c, _ := newClient()
		short, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		// Not covered by RequestTimeout: waits for caller context.
		_, err := c.API().HelpGetConfig(short)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("CallerContext", func(t *testing.T) {
		c, _ := newClient()
		short, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()
		// Caller deadline is earlier than request timeout: no resend.
		_, err := c.API().HelpGetNearestDC(short)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestClientRequestTimeoutTwice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := newTestClient(nil)
	c.conn = &hangConn{}
	c.ready.Signal()
	c.requestTimeout = func(bin.Encoder) time.Duration { return 20 * time.Millisecond }
	restarts := fakeReconnect(t, c, &hangConn{})

	_, err := c.API().HelpGetNearestDC(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "no response in")
	require.EqualValues(t, 1, restarts.Load(), "request is resent once")
}

func TestClientPingOptions(t *testing.T) {
	c := NewClient(1, "hash", Options{
		PingInterval: 5 * time.Second,
		PingTimeout:  2 * time.Second,
	})
	require.Equal(t, 5*time.Second, c.opts.PingInterval)
	require.Equal(t, 2*time.Second, c.opts.PingTimeout)
}
