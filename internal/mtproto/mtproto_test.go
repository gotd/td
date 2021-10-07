package mtproto

import (
	"context"
	"errors"
	"sync"

	"github.com/nnqq/td/bin"
)

type testPayload struct {
	Data []byte
}

func (d testPayload) Decode(b *bin.Buffer) error {
	_, err := b.Bytes()
	return err
}

func (d testPayload) Encode(b *bin.Buffer) error {
	b.PutBytes(d.Data)
	return nil
}

type noopBuf struct{}

func (n noopBuf) Consume(id int64) bool {
	return true
}

type constantConn struct {
	data    []byte
	cancel  context.CancelFunc
	counter int
	mux     sync.Mutex
}

func (c *constantConn) Send(ctx context.Context, b *bin.Buffer) error {
	return nil
}

func (c *constantConn) Recv(ctx context.Context, b *bin.Buffer) error {
	c.mux.Lock()
	exit := c.counter == 0
	if exit {
		c.mux.Unlock()
		c.cancel()
		return errors.New("error")
	}
	c.counter--
	c.mux.Unlock()

	b.Put(c.data)
	return nil
}

func (c *constantConn) Close() error {
	return nil
}
