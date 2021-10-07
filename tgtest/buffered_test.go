package tgtest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/transport"
)

func TestBufferedConn(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	i := transport.Intermediate
	c1, c2 := i.Pipe()
	b := newBufferedConn(c1)
	defer func() {
		a.NoError(b.Close())
		a.NoError(c2.Close())
	}()

	payload := []byte("abcdabcd")
	go func() {
		b1 := &bin.Buffer{Buf: payload}
		a.NoError(c2.Send(ctx, b1))
	}()

	// Test Recv before Push.
	recvBuf := &bin.Buffer{}
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(payload, recvBuf.Buf)

	pushed := []byte("12345678")
	b.Push(&bin.Buffer{Buf: pushed})
	go func() {
		b1 := &bin.Buffer{Buf: payload}
		a.NoError(c2.Send(ctx, b1))
	}()

	// Test Push.
	recvBuf.Reset()
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(pushed, recvBuf.Buf)

	// Test Recv after Push.
	recvBuf.Reset()
	a.NoError(b.Recv(ctx, recvBuf))
	a.Equal(payload, recvBuf.Buf)

	// Test send.
	go func() {
		b1 := &bin.Buffer{Buf: payload}
		a.NoError(b.Send(ctx, b1))
	}()

	recvBuf.Reset()
	a.NoError(c2.Recv(ctx, recvBuf))
	a.Equal(payload, recvBuf.Buf)
}
