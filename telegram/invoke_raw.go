package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tgerr"
)

// InvokeRaw sens input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Client) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := c.invokeRaw(ctx, input, output); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := tgerr.As(err); ok && rpcErr.IsCode(303) {
			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection.
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				c.log.Info("Got migrate error: Creating sub-connection",
					zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
				)
				return c.invokeSub(ctx, rpcErr.Argument, input, output)
			}

			c.log.Info("Got migrate error: Starting migration to another DC",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)
			// Otherwise we should change primary DC.
			return c.invokeMigrate(ctx, rpcErr.Argument, input, output)
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
