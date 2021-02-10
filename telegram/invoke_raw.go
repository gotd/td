package telegram

import (
	"context"

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
		if rpcErr, ok := mtproto.AsErr(err); ok && rpcErr.IsCode(303) {
			c.log.Info("Got migrate error: Starting migration to another dc",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)

			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection..
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				return c.invokeSub(ctx, rpcErr.Argument, input, output)
			}

			// Otherwise we should change primary DC.
			c.primaryDC.Store(int64(rpcErr.Argument))
			if err := c.migrateToDc(
				c.ctx, rpcErr.Argument,
				// TODO(tdakkota): Is it may be necessary to migrate if error is not FILE_MIGRATE or STATS_MIGRATE?
				false,
			); err != nil {
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
