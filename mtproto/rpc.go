package mtproto

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

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
		ID:       c.newMessageID(),
		Sequence: c.seqNo(true),
		Input:    input,
		Output:   output,
	}

	log := c.log.With(
		zap.Bool("content_msg", true),
		zap.Int64("msg_id", req.ID),
	)
	log.Debug("Invoke start")
	defer log.Debug("Invoke end")

	if err := c.rpc.Do(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug("Setting server salt")
			atomic.StoreInt64(&c.salt, badMsgErr.NewSalt)
			c.log.Info("Retrying request after basMsgErr", zap.Int64("msg_id", req.ID))
			return c.rpc.Do(ctx, req)
		}
		return xerrors.Errorf("rpcDoRequest: %w", err)
	}

	return nil
}

func (c *Conn) dropRPC(req rpc.Request) error {
	var resp mt.RpcDropAnswerBox

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := c.InvokeRaw(ctx, &mt.RPCDropAnswerRequest{
		ReqMsgID: req.ID,
	}, &resp); err != nil {
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
