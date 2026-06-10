package telegram

import (
	"context"
	"sync"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/rpc"
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
	}{
		{"ConnDead", errors.Wrap(pool.ErrConnDead, "waitSession")},
		{"EngineClosed", errors.Wrap(rpc.ErrEngineClosed, "engine forcibly closed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{log: zap.NewNop()}
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
	client := Client{log: zap.NewNop()}
	client.init()
	client.conn = newNotifyInvokeConn(pool.ErrConnDead)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// No reconnection happens, so canceled context must stop waiting,
	// reporting original invoke error.
	require.ErrorIs(t, client.invokeConn(ctx, nil, nil), pool.ErrConnDead)
}

func TestClient_invokeConnClientClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := Client{log: zap.NewNop(), ctx: ctx}
	client.init()
	client.conn = newNotifyInvokeConn(pool.ErrConnDead)

	// Closed client never reconnects, so waiting must stop immediately.
	require.ErrorIs(t, client.invokeConn(context.Background(), nil, nil), pool.ErrConnDead)
}

func TestClient_invokeConnNotRetryable(t *testing.T) {
	testErr := errors.New("some rpc error")

	client := Client{log: zap.NewNop()}
	client.init()
	client.conn = newNotifyInvokeConn(testErr)

	// Non connection-related errors must be returned as-is without waiting
	// for reconnect.
	require.ErrorIs(t, client.invokeConn(context.Background(), nil, nil), testErr)
}
