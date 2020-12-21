package telegram

import (
	"context"
	"errors"
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

	retryCtx, retryClose := context.WithCancel(ctx)
	defer retryClose()

	var (
		// Will write error to that channel.
		result = make(chan error)
		// Needed to prevent multiple handler calls.
		handlerCalls uint32
	)

	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		defer retryClose()

		atomic.AddUint32(&handlerCalls, 1)
		if atomic.LoadUint32(&handlerCalls) > 1 {
			log.Warn("handler already called")

			return xerrors.Errorf("handler already called")
		}

		if rpcErr != nil {
			result <- rpcErr
			return nil
		}

		decodeErr := req.Output.Decode(rpcBuff)
		result <- decodeErr
		return decodeErr
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
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case r := <-result:
		return r
	}
}

func (c *Client) rpcRetryUntilAck(ctx context.Context, req request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set ack callback for request.
	c.ackMux.Lock()
	c.ack[req.ID] = cancel
	c.ackMux.Unlock()

	defer func() {
		c.ackMux.Lock()
		delete(c.ack, req.ID)
		c.ackMux.Unlock()
	}()

	retries := 0
	for {
		select {
		case <-ctx.Done():
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
