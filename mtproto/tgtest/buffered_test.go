package tgtest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/transport"
)

func TestBufferedConn(t *testing.T) {
	a := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	i := transport.Intermediate(nil)
	c1, c2 := i.Pipe()

	payload := []byte("abcdabcd")
	go func() {
		b1 := &bin.Buffer{Buf: payload}
		a.NoError(c2.Send(ctx, b1))
	}()

	b := &BufferedConn{conn: c1}
	recvBuf := &bin.Buffer{}
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(payload, recvBuf.Buf)

	pushed := []byte("12345678")
	b.Push(&bin.Buffer{Buf: pushed})
	go func() {
		b1 := &bin.Buffer{Buf: payload}
		a.NoError(c2.Send(ctx, b1))
	}()

	recvBuf.Reset()
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(pushed, recvBuf.Buf)

	recvBuf.Reset()
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(payload, recvBuf.Buf)
}
