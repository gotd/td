package mtproto

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/tgerr"
)

func (c *Conn) handleResult(b *bin.Buffer) error {
	// Response to an RPC query.
	var res proto.Result
	if err := res.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}

	// Now b contains result message.
	b.ResetTo(res.Result)

	msgID := zap.Int64("msg_id", res.RequestMessageID)
	c.logWithBuffer(b).Debug("Handle result", msgID)

	// Handling gzipped results.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	if id == proto.GZIPTypeID {
		content, err := gzip(b)
		if err != nil {
			return xerrors.Errorf("decompress: %w", err)
		}

		// Replacing buffer so callback will deal with uncompressed data.
		b = content
		c.logWithBuffer(b).Debug("Decompressed", msgID)

		// Replacing id with inner id if error is compressed for any reason.
		if id, err = b.PeekID(); err != nil {
			return xerrors.Errorf("peek id: %w", err)
		}
	}

	if id == mt.RPCErrorTypeID {
		var rpcErr mt.RPCError
		if err := rpcErr.Decode(b); err != nil {
			return xerrors.Errorf("error decode: %w", err)
		}

		c.log.Debug("Got error", msgID,
			zap.Int("err_code", rpcErr.ErrorCode),
			zap.String("err_msg", rpcErr.ErrorMessage),
		)
		c.rpc.NotifyError(res.RequestMessageID, tgerr.New(rpcErr.ErrorCode, rpcErr.ErrorMessage))

		return nil
	}
	if id == mt.PongTypeID {
		return c.handlePong(b)
	}

	return c.rpc.NotifyResult(res.RequestMessageID, b)
}
