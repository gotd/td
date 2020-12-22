package telegram

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

func (c *Client) rpcDo(ctx context.Context, contentMsg bool, in bin.Encoder, out bin.Decoder) error {
	c.sessionCreated.WaitIfNeeded()

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
			if err := c.saveSession(c.ctx); err != nil {
				return xerrors.Errorf("badMsg update salt: %w", err)
			}

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

	retryCtx, retryClose := context.WithCancel(ctx)
	defer retryClose()

	// Will write error to that variable.
	var resultErr error
	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		defer func() { closeDone(); retryClose() }()

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
	if err := c.write(ctx, req.ID, req.Sequence, req.Input); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	// Start retrying.
	if err := c.rpcRetryUntilAck(retryCtx, req); err != nil {
		// If the retryCtx was canceled, then one of two things happened:
		// 1. User canceled the original context.
		// 2. The RPC result came and callback canceled retryCtx.
		//
		// If this is not an context.Canceled error, most likely we did not receive ACK
		// and exceeded the limit of attempts to send a request,
		// or could not write data to the connection, so we return an error.
		if !errors.Is(err, context.Canceled) {
			return xerrors.Errorf("retryUntilAck: %w", err)
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return resultErr
	}
}

// rpcRetryUntilAck resends the request to the server until ACK is received
// or context canceled.
//
// Returns nil if ACK was received, otherwise return error.
func (c *Client) rpcRetryUntilAck(ctx context.Context, req request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ackChan := make(chan error)
	go func() { ackChan <- c.waitACK(ctx, req.ID) }()

	retries := 0
	for {
		select {
		case ackErr := <-ackChan:
			if ackErr != nil {
				return xerrors.Errorf("wait ack: %w", ackErr)
			}

			return nil
			// TODO(ccln): use clock.
		case <-time.After(c.retryInterval):
			if err := c.write(ctx, req.ID, req.Sequence, req.Input); err != nil {
				c.log.Error("Retry attempt failed", zap.Error(err))
				return err
			}

			retries++
			if retries >= c.maxRetries {
				c.log.Error("Retry limit reached", zap.Int64("request_id", req.ID))
				return xerrors.New("retry limit reached")
			}
		}
	}
}
