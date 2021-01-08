package telegram

import (
	"context"
	"errors"

	"go.uber.org/zap"
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
		if errors.As(err, &rpcErr) && (rpcErr.Code == 303) {
			c.log.Info("Got migrate error: Starting migration to another dc",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)
			if err := c.migrateToDc(c.ctx, rpcErr.Argument); err != nil {
				return xerrors.Errorf("migrate to dc: %w", err)
			}
			// Re-trying request on another connection.
			return c.invokeRaw(ctx, input, output)
		}
		return err
	}
	return nil
}

func (c *Client) invokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c.connMux.Lock()
	conn := c.conn
	c.connMux.Unlock()

	return conn.InvokeRaw(ctx, input, output)
}
