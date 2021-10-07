package pool

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/tdsync"
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
