package tgtest

import (
	"context"
	"sync"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/transport"
)

type bufferedConn struct {
	conn transport.Conn

	recv    []bin.Buffer
	recvMux sync.Mutex
}

func newBufferedConn(conn transport.Conn) *bufferedConn {
	return &bufferedConn{conn: conn}
}

func (c *bufferedConn) push(b *bin.Buffer) {
	c.recvMux.Lock()
	c.recv = append(c.recv, bin.Buffer{Buf: b.Copy()})
	c.recvMux.Unlock()
}

func (c *bufferedConn) pop() (r bin.Buffer, ok bool) {
	c.recvMux.Lock()
	defer c.recvMux.Unlock()
	if len(c.recv) < 1 {
		return
	}
	r, c.recv = c.recv[len(c.recv)-1], c.recv[:len(c.recv)-1]
	ok = true
	return
}

func (c *bufferedConn) Push(b *bin.Buffer) {
	c.push(b)
}

func (c *bufferedConn) Pop() (bin.Buffer, bool) {
	return c.pop()
}

func (c *bufferedConn) Send(ctx context.Context, b *bin.Buffer) error {
	return c.conn.Send(ctx, b)
}

func (c *bufferedConn) Recv(ctx context.Context, b *bin.Buffer) error {
	e, ok := c.Pop()
	if ok {
		b.ResetTo(e.Copy())
		return nil
	}

	return c.conn.Recv(ctx, b)
}

func (c *bufferedConn) Close() error {
	return c.conn.Close()
}
