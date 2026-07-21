package pool

import (
	"context"
	"testing"

	"github.com/go-faster/errors"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/transport"
)

type mockConn struct {
	ready      *tdsync.Ready
	readyOnRun bool
}

func (mockConn) Ping(ctx context.Context) error {
	return errors.New("not implemented")
}

func newMockConn(readyOnRun bool) mockConn {
	return mockConn{
		ready:      tdsync.NewReady(),
		readyOnRun: readyOnRun,
	}
}

func (m mockConn) Run(ctx context.Context) error {
	if m.readyOnRun {
		m.ready.Signal()
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m mockConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return nil
}

func (m mockConn) Ready() <-chan struct{} {
	return m.ready.Ready()
}

type invokeErrConn struct {
	mockConn
	err error
}

func (m invokeErrConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return m.err
}

func TestDC_InvokeRetryOnDeadConn(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	created := 0
	p := NewDC(ctx, 2, func() Conn {
		created++
		if created == 1 {
			// First connection dies before request is sent.
			return invokeErrConn{
				mockConn: newMockConn(true),
				err:      errors.Wrap(ErrConnDead, "waitSession"),
			}
		}
		return newMockConn(true)
	}, DCOptions{
		MaxOpenConnections: 1,
	})
	defer func() {
		a.NoError(p.Close())
	}()

	// Request must be transparently retried on a new connection.
	a.NoError(p.Invoke(ctx, nil, nil))
	a.Equal(2, created, "Pool must create new connection to retry invoke")
}

func TestDC_InvokeRetryOnWriteFailed(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	created := 0
	p := NewDC(ctx, 2, func() Conn {
		created++
		if created == 1 {
			// First connection's write failed: the frame was not sent, or
			// sent partially and discarded by the server, so retrying on a
			// new connection is safe.
			return invokeErrConn{
				mockConn: newMockConn(true),
				err:      errors.Wrap(transport.ErrWriteFailed, "write"),
			}
		}
		return newMockConn(true)
	}, DCOptions{
		MaxOpenConnections: 1,
	})
	defer func() {
		a.NoError(p.Close())
	}()

	// Request must be transparently retried on a new connection.
	a.NoError(p.Invoke(ctx, nil, nil))
	a.Equal(2, created, "Pool must create new connection to retry invoke")
}

func TestDC_InvokeNotRetryable(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	testErr := errors.New("some rpc error")

	created := 0
	p := NewDC(ctx, 2, func() Conn {
		created++
		return invokeErrConn{
			mockConn: newMockConn(true),
			err:      testErr,
		}
	}, DCOptions{
		MaxOpenConnections: 1,
	})
	defer func() {
		a.NoError(p.Close())
	}()

	// Non connection-related errors must be returned as-is.
	a.ErrorIs(p.Invoke(ctx, nil, nil), testErr)
	a.Equal(1, created, "Pool must not retry on non-retryable error")
}

func TestDC_acquire(t *testing.T) {
	t.Run("AcquireRelease", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()

		created := 0
		p := NewDC(ctx, 2, func() Conn {
			created++
			return newMockConn(true)
		}, DCOptions{
			MaxOpenConnections: 1,
		})
		defer func() {
			a.NoError(p.Close())
		}()

		c, err := p.acquire(ctx)
		a.NoError(err)
		a.NotNil(c)
		a.Equal(1, created, "Pool must create new connection")

		p.release(c)

		_, err = p.acquire(ctx)
		a.NoError(err)
		a.Equal(1, created, "Pool must re-use connection")

		p.release(c)
	})
	t.Run("CancelWhileWait", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()

		created := 0
		p := NewDC(ctx, 2, func() Conn {
			created++
			return newMockConn(true)
		}, DCOptions{
			MaxOpenConnections: 1,
		})
		defer func() {
			a.NoError(p.Close())
		}()

		c, err := p.acquire(ctx)
		a.NoError(err)
		a.NotNil(c)
		a.Equal(1, created, "Pool must create new connection")

		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		c2, err := p.acquire(canceledCtx)
		a.ErrorIs(err, context.Canceled)
		a.Nil(c2)
		a.Empty(p.freeReq.m)
	})
	t.Run("Dead", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()

		created := 0
		p := NewDC(ctx, 2, func() Conn {
			created++
			return newMockConn(true)
		}, DCOptions{
			MaxOpenConnections: 1,
		})
		defer func() {
			a.NoError(p.Close())
		}()

		c, err := p.acquire(ctx)
		a.NoError(err)
		a.NotNil(c)
		a.Equal(1, created, "Pool must create new connection")

		p.release(c)
		c.dead.Signal()

		_, err = p.acquire(ctx)
		a.NoError(err)
		a.Equal(2, created, "Pool must not re-use dead connection")
	})
	t.Run("CancelWhileCreate", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()

		created := 0
		p := NewDC(ctx, 2, func() Conn {
			created++
			return newMockConn(false)
		}, DCOptions{
			MaxOpenConnections: 1,
		})
		defer func() {
			a.NoError(p.Close())
		}()

		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		c, err := p.acquire(canceledCtx)
		a.ErrorIs(err, context.Canceled)
		a.Nil(c)
	})
}
