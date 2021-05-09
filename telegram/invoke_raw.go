package telegram

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tgerr"
)

// InvokeRaw sends input and decodes result into output.
func (c *Client) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return c.invoker.InvokeRaw(ctx, input, output)
}

// clientInvoker implements tg.Invoker on Client without middleware support.
type clientInvoker struct {
	*Client
}

// InvokeRaw sends input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c clientInvoker) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := c.connInvokeRaw(ctx, input, output); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := tgerr.As(err); ok && rpcErr.IsCode(303) {
			targetDC := rpcErr.Argument
			log := c.log.With(
				zap.String("error_type", rpcErr.Type),
				zap.Int("target_dc", targetDC),
			)
			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection.
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				log.Debug("Invoking on target DC")
				return c.invokeSub(ctx, targetDC, input, output)
			}

			// Otherwise we should change primary DC.
			log.Info("Migrating to target DC")
			return c.invokeMigrate(ctx, targetDC, input, output)
		}

		return err
	}
	return nil
}

func (c *Client) connInvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c.connMux.Lock()
	conn := c.conn
	c.connMux.Unlock()

	return conn.InvokeRaw(ctx, input, output)
}
