package telegram

import (
	"context"
	"sync"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
)

func (c *Client) do(ctx context.Context, req bin.Encoder, res bin.Decoder) error {
	id := c.newMessageID()

	// Creating "done" channel and ensuring that it will be closed before
	// method returns and only once.
	done := make(chan struct{})
	doneClose := sync.Once{}
	closeDone := func() {
		doneClose.Do(func() {
			close(done)
		})
	}
	defer closeDone()

	// Will write error to that variable.
	var resultErr error

	// Setting callback that will be called if message is received.
	c.rpcMux.Lock()
	c.rpc[id] = func(rpcBuff *bin.Buffer, rpcErr error) {
		if rpcErr != nil {
			// Not calling f, just returning error.
			resultErr = rpcErr
		} else {
			resultErr = res.Decode(rpcBuff)
		}
		closeDone()
	}
	c.rpcMux.Unlock()

	defer func() {
		// Ensuring that callback can't be called after function return.
		c.rpcMux.Lock()
		delete(c.rpc, id)
		c.rpcMux.Unlock()
	}()

	// Encoding request. Note that callback is already set.
	if err := c.write(id, req); err != nil {
		return xerrors.Errorf("failed to write: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return resultErr
	}
}
