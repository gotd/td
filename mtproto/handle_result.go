package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tgerr"
)

func (c *Conn) handleResult(b *bin.Buffer) error {
	ctx := context.Background()
	// Response to an RPC query.
	var res proto.Result
	if err := res.Decode(b); err != nil {
		return errors.Wrap(err, "decode")
	}

	// Now b contains result message.
	b.ResetTo(res.Result)

	msgID := log.Int64("msg_id", res.RequestMessageID)
	c.logWithBuffer(b).Debug(ctx, "Handle result", msgID)

	// Handling gzipped results.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	if id == proto.GZIPTypeID {
		content, err := gzip(b)
		if err != nil {
			return errors.Wrap(err, "decompress")
		}

		// Replacing buffer so callback will deal with uncompressed data.
		b = content
		c.logWithBuffer(b).Debug(ctx, "Decompressed", msgID)

		// Replacing id with inner id if error is compressed for any reason.
		if id, err = b.PeekID(); err != nil {
			return errors.Wrap(err, "peek id")
		}
	}

	if id == mt.RPCErrorTypeID {
		var rpcErr mt.RPCError
		if err := rpcErr.Decode(b); err != nil {
			return errors.Wrap(err, "error decode")
		}

		c.log.Debug(ctx, "Got error", msgID,
			log.Int("err_code", rpcErr.ErrorCode),
			log.String("err_msg", rpcErr.ErrorMessage),
		)
		if c.pfs && rpcErr.ErrorCode == 401 && rpcErr.ErrorMessage == "AUTH_KEY_PERM_EMPTY" {
			// The server forgot the binding of the temporary key. Like TDLib
			// (and Telegram Android/Desktop), drop the temporary key and
			// reconnect: the request fails as sent on a dead connection, so
			// callers resend it on the new one after re-bind.
			c.log.Warn(ctx, "Temporary key binding lost (AUTH_KEY_PERM_EMPTY), reconnecting", msgID)
			c.rpc.NotifyError(res.RequestMessageID,
				errors.Wrap(rpc.ErrEngineClosed, "temporary key binding lost (AUTH_KEY_PERM_EMPTY)"))
			c.pfsBindingLost.Signal()
			return nil
		}
		c.rpc.NotifyError(res.RequestMessageID, tgerr.New(rpcErr.ErrorCode, rpcErr.ErrorMessage))

		return nil
	}
	if id == mt.PongTypeID {
		return c.handlePong(b)
	}

	return c.rpc.NotifyResult(res.RequestMessageID, b)
}
