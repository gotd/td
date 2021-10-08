package telegram

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

// API returns *tg.Client for calling raw MTProto methods.
func (c *Client) API() *tg.Client {
	return c.tg
}

// Invoke invokes raw MTProto RPC method. It sends input and decodes result
// into output.
func (c *Client) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return c.invoker.Invoke(ctx, input, output)
}

// invokeDirect directly invokes RPC method, automatically handling datacenter redirects.
func (c *Client) invokeDirect(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := c.invokeConn(ctx, input, output); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := tgerr.As(err); ok && strings.HasSuffix(rpcErr.Type, "_MIGRATE") {
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

// invokeConn directly invokes RPC call on primary connection.
func (c *Client) invokeConn(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c.connMux.Lock()
	conn := c.conn
	c.connMux.Unlock()

	err := conn.Invoke(ctx, input, output)
	if err != nil {
		e := err.Error()
		if strings.Contains(e, "engine was closed") ||
			strings.Contains(e, "broken pipe") ||
			strings.Contains(e, "i/o timeout") {
			log := c.log.With(zap.String("restart_reason", e))

			log.Info("got network error, begin restart")
			c.restart <- struct{}{}

			<-c.restarted
			log.Info("restarted, retry invoke")

			c.connMux.Lock()
			newConn := c.conn
			c.connMux.Unlock()
			return newConn.Invoke(ctx, input, output)
		}

		return err
	}

	return nil
}
