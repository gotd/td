package telegram

import (
	"context"
	"sync"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/transport"
)

// notifyInvokeConn is a fake connection which always fails Invoke with given
// error, signaling about every Invoke call via channel.
type notifyInvokeConn struct {
	err     error
	once    sync.Once
	invoked chan struct{}
}

func newNotifyInvokeConn(err error) *notifyInvokeConn {
	return &notifyInvokeConn{
		err:     err,
		invoked: make(chan struct{}),
	}
}

func (c *notifyInvokeConn) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (c *notifyInvokeConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	c.once.Do(func() { close(c.invoked) })
	return c.err
}

func (c *notifyInvokeConn) Ping(context.Context) error { return nil }

type okInvokeConn struct{}

func (okInvokeConn) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (okInvokeConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error { return nil }

func (okInvokeConn) Ping(context.Context) error { return nil }

func TestClient_invokeConnRetriesOnNewConn(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		// Retrying a failed transport send is opt-in; the other two classes
		// are retried unconditionally.
		retryOnWriteFailed bool
	}{
		{name: "ConnDead", err: errors.Wrap(pool.ErrConnDead, "waitSession")},
		{name: "EngineClosed", err: errors.Wrap(rpc.ErrEngineClosed, "engine forcibly closed")},
		{name: "WriteFailed", err: errors.Wrap(transport.ErrWriteFailed, "write"), retryOnWriteFailed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{log: log.For(log.Nop), retryOnWriteFailed: tt.retryOnWriteFailed}
			client.init()

			dead := newNotifyInvokeConn(tt.err)
			client.conn = dead

			// Replace dead connection like reconnection loop does, but only
			// after invokeConn observed the dead connection.
			go func() {
				<-dead.invoked
				client.connMux.Lock()
				client.replaceConn(okInvokeConn{})
				client.connMux.Unlock()
			}()

			// Request must be transparently retried on the new connection.
			require.NoError(t, client.invokeConn(context.Background(), nil, nil))
		})
	}
}

func TestClient_invokeConnContextCancel(t *testing.T) {
	client := Client{log: log.For(log.Nop)}
	client.init()
	client.conn = newNotifyInvokeConn(pool.ErrConnDead)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// No reconnection happens, so canceled context must stop waiting,
	// reporting context error to keep standard context semantics.
	err := client.invokeConn(ctx, nil, nil)
	require.ErrorIs(t, err, context.Canceled)
	require.NotErrorIs(t, err, pool.ErrConnDead)
}

func TestClient_invokeConnClientClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := Client{log: log.For(log.Nop), ctx: ctx}
	client.init()
	client.conn = newNotifyInvokeConn(pool.ErrConnDead)

	// Closed client never reconnects, so waiting must stop immediately,
	// reporting client context error.
	err := client.invokeConn(context.Background(), nil, nil)
	require.ErrorIs(t, err, context.Canceled)
	require.NotErrorIs(t, err, pool.ErrConnDead)
}

func TestErrRetryableOnNewConn_AckedCloseNotRetryable(t *testing.T) {
	// An acknowledged request may already have been processed by the
	// server: errRetryableOnNewConn must NOT classify it as safe to retry
	// on a new connection, or a transparent resend could duplicate the RPC.
	// This holds regardless of the write-failure opt-in.
	require.False(t, errRetryableOnNewConn(rpc.ErrEngineClosedAfterAck, false))
	require.False(t, errRetryableOnNewConn(rpc.ErrEngineClosedAfterAck, true))
}

func TestClient_invokeConnWriteFailedSurfacesByDefault(t *testing.T) {
	// Retrying a failed transport send is opt-in. By default the error is
	// returned to the caller, who may be acting on it — rotating a proxy or
	// an endpoint — rather than being retried transparently.
	client := Client{log: log.For(log.Nop)}
	client.init()
	client.conn = newNotifyInvokeConn(errors.Wrap(transport.ErrWriteFailed, "write"))

	err := client.invokeConn(context.Background(), nil, nil)
	require.ErrorIs(t, err, transport.ErrWriteFailed)
}

func TestClient_invokeConnNotRetryable(t *testing.T) {
	testErr := errors.New("some rpc error")

	client := Client{log: log.For(log.Nop)}
	client.init()
	client.conn = newNotifyInvokeConn(testErr)

	// Non connection-related errors must be returned as-is without waiting
	// for reconnect.
	require.ErrorIs(t, client.invokeConn(context.Background(), nil, nil), testErr)
}
