package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
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
		var content proto.GZIP
		if err := content.Decode(b); err != nil {
			return xerrors.Errorf("failed to decode: %w", err)
		}
		// Replacing buffer so callback will deal with uncompressed data.
		b = &bin.Buffer{Buf: content.Data}
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
