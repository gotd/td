package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/rpc"
)

// Invoke sends input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Conn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	msgID, seqNo := c.nextMsgSeq(true)
	req := rpc.Request{
		MsgID:  msgID,
		SeqNo:  seqNo,
		Input:  input,
		Output: output,
	}

	log := c.log.With(
		zap.Int64("msg_id", req.MsgID),
	)
	log.Debug("Invoke start")
	defer log.Debug("Invoke end")

	if err := c.rpc.Do(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug("Setting server salt")
			// Store salt from server.
			c.storeSalt(badMsgErr.NewSalt)
			// Reset saved salts to fetch new.
			c.salts.Reset()
			c.log.Info("Retrying request after basMsgErr", zap.Int64("msg_id", req.MsgID))
			return c.rpc.Do(ctx, req)
		}
		return errors.Wrap(err, "rpcDoRequest")
	}

	return nil
}

func (c *Conn) dropRPC(req rpc.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		c.getTimeout(mt.RPCDropAnswerRequestTypeID),
	)
	defer cancel()

	var resp mt.RPCDropAnswerBox
	if err := c.Invoke(ctx, &mt.RPCDropAnswerRequest{
		ReqMsgID: req.MsgID,
	}, &resp); err != nil {
		return err
	}

	switch resp.RpcDropAnswer.(type) {
	case *mt.RPCAnswerDropped, *mt.RPCAnswerDroppedRunning:
		return nil
	case *mt.RPCAnswerUnknown:
		return errors.New("answer unknown")
	default:
		return errors.Errorf("unexpected response type: %T", resp.RpcDropAnswer)
	}
}
