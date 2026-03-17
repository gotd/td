package telegram

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
)

type runClientConn struct {
	run func(ctx context.Context) error
}

func (r runClientConn) Run(ctx context.Context) error {
	return r.run(ctx)
}

func (runClientConn) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (runClientConn) Ping(context.Context) error {
	return nil
}

type cancelCheckInvoker struct {
	client *Client

	closed              atomic.Bool
	canceledBeforeClose atomic.Bool
}

type inProgressCloseInvoker struct {
	started chan struct{}
	unblock chan struct{}
	done    chan struct{}

	calls atomic.Int32
	once  sync.Once
}

func (*inProgressCloseInvoker) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (c *inProgressCloseInvoker) Close() error {
	if c.calls.Add(1) > 1 {
		return errors.New("DC already closed")
	}

	c.once.Do(func() {
		close(c.started)
	})
	<-c.unblock
	close(c.done)
	return nil
}

func (*cancelCheckInvoker) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (c *cancelCheckInvoker) Close() error {
	c.closed.Store(true)

	if c.client != nil && c.client.ctx != nil {
		select {
		case <-c.client.ctx.Done():
			c.canceledBeforeClose.Store(true)
		default:
			c.canceledBeforeClose.Store(false)
		}
	}

	return nil
}

func TestClientRunCancelsContextBeforeClosingManagedConns(t *testing.T) {
	c := NewClient(1, "hash", Options{
		NoUpdates: true,
		Logger:    zap.NewNop(),
	})

	checker := &cancelCheckInvoker{client: c}
	c.subConns[1] = checker
	c.conn = runClientConn{
		run: func(ctx context.Context) error {
			c.onReady()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	err := c.Run(context.Background(), func(context.Context) error { return nil })
	require.NoError(t, err)
	require.True(t, checker.closed.Load())
	require.True(t, checker.canceledBeforeClose.Load())
}

func TestClientRunSkipsDoubleCloseForAlreadyClosingCDNConn(t *testing.T) {
	c := NewClient(1, "hash", Options{
		NoUpdates: true,
		Logger:    zap.NewNop(),
	})

	inv := &inProgressCloseInvoker{
		started: make(chan struct{}),
		unblock: make(chan struct{}),
		done:    make(chan struct{}),
	}
	c.cdnPools.conns[203] = []cachedCDNPool{{
		conn: inv,
		max:  1,
	}}
	c.cdnPools.invalidateDC(203)

	select {
	case <-inv.started:
	case <-time.After(time.Second):
		t.Fatal("expected close worker to start")
	}

	c.conn = runClientConn{
		run: func(ctx context.Context) error {
			c.onReady()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	err := c.Run(context.Background(), func(context.Context) error { return nil })
	require.NoError(t, err)
	require.EqualValues(t, 1, inv.calls.Load(), "shutdown must not issue second close")

	close(inv.unblock)
	select {
	case <-inv.done:
	case <-time.After(time.Second):
		t.Fatal("expected close worker to finish")
	}
}
