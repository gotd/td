package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/mt"
	"github.com/ernado/td/internal/proto"
)

// Error represents RPC error returned to request.
type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("rpc error code %d: %s", e.Code, e.Message)
}

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
		var content proto.GZIP
		if err := content.Decode(b); err != nil {
			return xerrors.Errorf("failed to decode: %w", err)
		}
		// Replacing buffer so callback will deal with uncompressed data.
		b = &bin.Buffer{Buf: content.Data}

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
