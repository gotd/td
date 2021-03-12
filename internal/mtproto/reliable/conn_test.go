package reliable_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/lifetime"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/mtproto/reliable"
	"github.com/stretchr/testify/require"
)

type UnstableMTConn struct {
	fatal chan error
	onReq func()
}

func NewUnstableConn(onRequest func()) *UnstableMTConn {
	return &UnstableMTConn{
		fatal: make(chan error),
		onReq: onRequest,
	}
}

func (c *UnstableMTConn) Run(ctx context.Context, f func(context.Context) error) error {
	echan := make(chan error)
	go func() { echan <- f(ctx) }()
	go func() { echan <- <-c.fatal }()
	err := <-echan
	return err
}

func (c *UnstableMTConn) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	c.onReq()
	return nil
}

func (c *UnstableMTConn) Break(e error) {
	c.fatal <- e
}

func TestConn(t *testing.T) {
	var (
		conn             *UnstableMTConn
		createConnCalls  = 0
		onConnectedCalls = 0
		onConnected      = make(chan struct{}, 10)
		requests         = 0
		connected        = false
	)

	reli := reliable.New(reliable.Config{
		CreateConn: func(addr string, opts mtproto.Options) reliable.MTConn {
			if connected {
				t.Fatal("multiple create conn calls")
			}
			createConnCalls++
			return NewUnstableConn(func() { requests++ })
		},
		OnConnected: func(m reliable.MTConn) error {
			if !connected {
				onConnectedCalls++
				connected = true
				onConnected <- struct{}{}
				conn = m.(*UnstableMTConn)
				return nil
			}
			t.Fatal("multiple onConnected calls")
			return nil
		},
	})

	life, err := lifetime.Start(reli)
	require.NoError(t, err)

	<-onConnected
	require.True(t, connected)
	require.Equal(t, 1, createConnCalls)
	require.Equal(t, 1, onConnectedCalls)
	require.NotNil(t, conn)

	require.NoError(t, reli.InvokeRaw(context.TODO(), nil, nil))
	require.NoError(t, reli.InvokeRaw(context.TODO(), nil, nil))
	require.Equal(t, 2, requests)

	connected = false
	conn.Break(fmt.Errorf("break error"))

	<-onConnected
	require.True(t, connected)
	require.Equal(t, 2, createConnCalls)
	require.Equal(t, 2, onConnectedCalls)
	require.NotNil(t, conn)

	require.NoError(t, reli.InvokeRaw(context.TODO(), nil, nil))
	require.NoError(t, reli.InvokeRaw(context.TODO(), nil, nil))
	require.Equal(t, 4, requests)

	require.NoError(t, life.Stop())
}
