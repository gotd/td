package tgtest

import (
	"context"
	"sync"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/transport"
)

type BufferedConn struct {
	conn transport.Conn

	recv    []bin.Buffer
	recvMux sync.Mutex
}

func NewBufferedConn(conn transport.Conn) *BufferedConn {
	return &BufferedConn{conn: conn}
}

func (c *BufferedConn) push(b *bin.Buffer) {
	c.recvMux.Lock()
	c.recv = append(c.recv, bin.Buffer{Buf: b.Copy()})
	c.recvMux.Unlock()
}

func (c *BufferedConn) pop() (r bin.Buffer, ok bool) {
	c.recvMux.Lock()
	defer c.recvMux.Unlock()
	if len(c.recv) < 1 {
		return
	}
	r, c.recv = c.recv[len(c.recv)-1], c.recv[:len(c.recv)-1]
	ok = true
	return
}

func (c *BufferedConn) Push(b *bin.Buffer) {
	c.push(b)
}

func (c *BufferedConn) Send(ctx context.Context, b *bin.Buffer) error {
	return c.conn.Send(ctx, b)
}

func (c *BufferedConn) Recv(ctx context.Context, b *bin.Buffer) error {
	e, ok := c.pop()
	if ok {
		b.ResetTo(e.Copy())
		return nil
	}

	return c.conn.Recv(ctx, b)
}

func (c *BufferedConn) Close() error {
	return c.conn.Close()
}
