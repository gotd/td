package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/mt"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) handleResult(b *bin.Buffer) error {
	// Response to an RPC query.
	var res proto.Result
	if err := res.Decode(b); err != nil {
		return xerrors.Errorf("failed to decode: %x", err)
	}
	c.log.With(
		zap.Int64("request_id", int64(res.RequestMessageID)),
	).Debug("Handle result")

	// Handling gzipped results.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	if id == proto.GZIPTypeID {
		content, err := c.gzip(b)
		if err != nil {
			return xerrors.Errorf("failed to decompres: %w", err)
		}

		// Replacing buffer so callback will deal with uncompressed data.
		b = content

		// Replacing id with inner id if error is compressed for any reason.
		if id, err = b.PeekID(); err != nil {
			return xerrors.Errorf("failed to peek id: %w", err)
		}
	}

	if id == mt.RPCErrorTypeID {
		var rpcErr mt.RPCError
		if err := rpcErr.Decode(b); err != nil {
			return xerrors.Errorf("failed to decode: %w", err)
		}

		c.rpcMux.Lock()
		f, ok := c.rpc[res.RequestMessageID]
		c.rpcMux.Unlock()
		if ok {
			f(nil, &Error{Code: rpcErr.ErrorCode, Message: rpcErr.ErrorMessage})
		}

		return nil
	}
	if id == mt.PongTypeID {
		return c.handlePong(b)
	}

	c.rpcMux.Lock()
	f, ok := c.rpc[res.RequestMessageID]
	c.rpcMux.Unlock()

	if ok {
		f(b, nil)
	} else {
		c.log.Debug("Got unexpected result")
	}

	return nil
}
