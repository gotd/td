package telegram

import (
	"context"
	"errors"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
)

// InvokeRaw sens input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Client) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := c.invokeRaw(ctx, input, output); err != nil {
		// Handling datacenter migration request.
		var rpcErr *mtproto.Error
		if errors.As(err, &rpcErr) && rpcErr.Type == "USER_MIGRATE" {
			if err := c.migrateToDc(c.ctx, rpcErr.Argument); err != nil {
				return xerrors.Errorf("migrate to dc: %w", err)
			}
			// Re-trying request on another connection.
			return c.invokeRaw(ctx, input, output)
		}
	}
	return nil
}

func (c *Client) invokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c.connMux.Lock()
	defer c.connMux.Unlock()
	return c.conn.InvokeRaw(ctx, input, output)
}
