package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
)

func (c *Client) handleResult(b *bin.Buffer) error {
	// Response to an RPC query.
	var res proto.Result
	if err := res.Decode(b); err != nil {
		return xerrors.Errorf("decode: %x", err)
	}
	c.log.With(
		zap.Int64("request_id", res.RequestMessageID),
	).Debug("Handle result")

	// Handling gzipped results.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	if id == proto.GZIPTypeID {
		content, err := c.gzip(b)
		if err != nil {
			return xerrors.Errorf("decompress: %w", err)
		}

		// Replacing buffer so callback will deal with uncompressed data.
		b = content

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

		c.rpcMux.Lock()
		f, ok := c.rpc[res.RequestMessageID]
		c.rpcMux.Unlock()
		if ok {
			e := &Error{
				Code:    rpcErr.ErrorCode,
				Message: rpcErr.ErrorMessage,
			}
			e.extractArgument()
			return f(nil, e)
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
		return f(b, nil)
	}

	c.log.Debug("Got unexpected result")
	return nil
}
