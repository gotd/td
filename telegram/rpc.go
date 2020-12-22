package telegram

import (
	"context"
	"errors"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/rpc"
)

func (c *Client) rpcDo(ctx context.Context, contentMsg bool, in bin.Encoder, out bin.Decoder) error {
	c.sessionCreated.WaitIfNeeded()

	req := rpc.Request{
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

	if err := c.rpc.Do(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug("Setting server salt")
			atomic.StoreInt64(&c.salt, badMsgErr.NewSalt)
			if err := c.saveSession(c.ctx); err != nil {
				return xerrors.Errorf("badMsg update salt: %w", err)
			}
			c.log.Info("Retrying request after basMsgErr", zap.Int64("msg_id", req.ID))
			return c.rpc.Do(ctx, req)
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
