package telegram

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

func (c *Client) rpcDo(ctx context.Context, contentMsg bool, in bin.Encoder, out bin.Decoder) error {
	req := request{
		ID:     c.newMessageID(),
		Input:  in,
		Output: out,
	}

	c.sentContentMessagesMux.Lock()
	// Atomically calculating and updating sequence number.
	req.Sequence = c.sentContentMessages * 2
	if contentMsg {
		req.Sequence++
		c.sentContentMessages++
	}
	c.sentContentMessagesMux.Unlock()

	if err := c.rpcDoRequest(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug("Setting server salt")
			atomic.StoreInt64(&c.salt, badMsgErr.NewSalt)
			return c.rpcDoRequest(ctx, req)
		}
		return xerrors.Errorf("rpcDoRequest: %w", err)
	}

	return nil
}

// rpcContent should be called for RPC requests that require acknowledgement, like
// content requests (send message, etc).
func (c *Client) rpcContent(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	return c.rpcDo(ctx, true, in, out)
}

type request struct {
	ID       int64
	Sequence int32
	Input    bin.Encoder
	Output   bin.Decoder
}

// rpcDoRequest starts an RPC request, setting handler for result and sending
// it to Telegram server.
func (c *Client) rpcDoRequest(ctx context.Context, req request) error {
	log := c.log.With(zap.Int("message_id", int(req.ID)))
	log.Debug("Do")

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
	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		defer closeDone()

		if rpcErr != nil {
			resultErr = rpcErr
			return nil
		}

		resultErr = req.Output.Decode(rpcBuff)
		return resultErr
	}

	// Setting callback that will be called if message is received.
	c.rpcMux.Lock()
	c.rpc[req.ID] = handler
	c.rpcMux.Unlock()

	defer func() {
		// Ensuring that callback can't be called after function return.
		c.rpcMux.Lock()
		delete(c.rpc, req.ID)
		c.rpcMux.Unlock()
	}()

	// Encoding request. Note that callback is already set.
	if err := c.write(req.ID, req.Sequence, req.Input); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	ackCtx, ackClose := context.WithCancel(c.ctx)
	defer ackClose()

	// resend request until we receive ack or response for it
	go c.ackOutcomingRPC(ackCtx, req)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return resultErr
	}
}
