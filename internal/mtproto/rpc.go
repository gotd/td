package mtproto

import (
	"context"
	"errors"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/rpc"
)

// InvokeRaw sends input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Conn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	req := rpc.Request{
		ID:     atomic.AddInt64(&c.reqID, 1),
		Input:  input,
		Output: output,
	}

	log := c.log.With(
		zap.Int64("req_id", req.ID),
	)
	log.Debug("Invoke start")
	defer log.Debug("Invoke end")

	defer c.cleanup(req.ID)

	if err := c.rpc.Do(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug("Setting server salt")
			// Store salt from server.
			c.storeSalt(badMsgErr.NewSalt)
			// Reset saved salts to fetch new.
			c.salts.Reset()
			c.log.Info("Retrying request after basMsgErr", zap.Int64("req_id", req.ID))
			return c.rpc.Do(ctx, req)
		}
		return xerrors.Errorf("rpcDoRequest: %w", err)
	}

	return nil
}

func (c *Conn) dropRPC(req rpc.Request) error {
	var resp mt.RPCDropAnswerBox

	dropReq := &mt.RPCDropAnswerRequest{
		ReqMsgID: req.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		c.getTimeout(dropReq.TypeID()),
	)
	defer cancel()

	if err := c.InvokeRaw(ctx, dropReq, &resp); err != nil {
		return err
	}

	switch resp.RpcDropAnswer.(type) {
	case *mt.RPCAnswerDropped, *mt.RPCAnswerDroppedRunning:
		return nil
	case *mt.RPCAnswerUnknown:
		return xerrors.Errorf("unknown request id")
	default:
		return xerrors.Errorf("unexpected response type: %T", resp.RpcDropAnswer)
	}
}
