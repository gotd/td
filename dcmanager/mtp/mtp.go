package mtp

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"golang.org/x/xerrors"
)

type Conn struct {
	conn *mtproto.Conn

	closeCh   chan struct{}
	runResult chan error

	err    error
	closed int32
}

func New(addr string, opts mtproto.Options) (*Conn, error) {
	wrapper := &Conn{
		conn:      mtproto.New(addr, opts),
		closeCh:   make(chan struct{}),
		runResult: make(chan error),
	}

	connected := make(chan struct{})
	go func() {
		wrapper.runResult <- wrapper.conn.Run(context.Background(), func(ctx context.Context) error {
			close(connected)
			<-wrapper.closeCh
			return context.Canceled
		})
	}()

	select {
	case err := <-wrapper.runResult:
		return nil, err
	case <-connected:
		return wrapper, nil
	case <-time.After(time.Second * 20):
		return nil, xerrors.Errorf("timeout")
	}
}

func (c *Conn) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	if ok, err := c.isClosed(); ok {
		return err
	}

	return c.conn.InvokeRaw(ctx, in, out)
}

func (c *Conn) Close() error {
	if ok, err := c.isClosed(); ok {
		return xerrors.Errorf("already closed: %w", err)
	}

	close(c.closeCh)
	err := <-c.runResult
	c.setError(err)
	return err
}

func (c *Conn) setError(err error) {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.err = err
	}
}

func (c *Conn) isClosed() (bool, error) {
	if atomic.LoadInt32(&c.closed) == 1 {
		return true, c.err
	}

	return false, nil
}
