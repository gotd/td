package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// API returns *tg.Client for calling raw MTProto methods.
func (c *Client) API() *tg.Client {
	return c.tg
}

// Invoke invokes raw MTProto RPC method. It sends input and decodes result
// into output.
func (c *Client) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if c.tracer != nil {
		spanName := "Invoke"
		var attrs []attribute.KeyValue
		if t, ok := input.(interface{ TypeID() uint32 }); ok {
			id := t.TypeID()
			attrs = append(attrs,
				attribute.Int64("tg.method.id_int", int64(id)),
				attribute.String("tg.method.id", fmt.Sprintf("%x", id)),
			)
			name := c.opts.Types.Get(id)
			if name == "" {
				name = fmt.Sprintf("0x%x", id)
			} else {
				attrs = append(attrs, attribute.String("tg.method.name", name))
			}
			spanName = fmt.Sprintf("Invoke: %s", name)
		}
		spanCtx, span := c.tracer.Start(ctx, spanName,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		ctx = spanCtx
		defer span.End()
	}

	return c.invoker.Invoke(ctx, input, output)
}

// invokeDirect directly invokes RPC method, automatically handling datacenter redirects.
func (c *Client) invokeDirect(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := c.invokeConn(ctx, input, output); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := tgerr.As(err); ok && strings.HasSuffix(rpcErr.Type, "_MIGRATE") {
			targetDC := rpcErr.Argument
			logger := c.log.With(
				log.String("error_type", rpcErr.Type),
				log.Int("target_dc", targetDC),
			)
			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection.
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				logger.Debug(ctx, "Invoking on target DC")
				return c.invokeSub(ctx, targetDC, input, output)
			}

			// Otherwise we should change primary DC.
			logger.Info(ctx, "Migrating to target DC")
			return c.invokeMigrate(ctx, targetDC, input, output)
		}

		return err
	}

	return nil
}

// invokeConn directly invokes RPC call on primary connection without any
// additional handling.
//
// If the connection dies before the request is processed by the server,
// invokeConn waits until the reconnection loop replaces the connection and
// retries the request on it, see https://github.com/gotd/td/issues/1030.
func (c *Client) invokeConn(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	attempt := 0
	for {
		c.connMux.Lock()
		conn := c.conn
		connChanged := c.connChanged
		c.connMux.Unlock()

		err := conn.Invoke(ctx, input, output)
		if err == nil || !errRetryableOnNewConn(err) {
			return err
		}

		var clientDone <-chan struct{}
		if c.ctx != nil {
			clientDone = c.ctx.Done()
		}
		attempt++
		c.log.Debug(ctx, "Primary connection is dead, waiting for new connection to retry",
			log.Error(err),
			log.Int("attempt", attempt),
		)
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "wait for reconnect")
		case <-clientDone:
			// Client is closed, no reconnection will happen.
			return errors.Wrap(c.ctx.Err(), "client closed")
		case <-connChanged:
			c.log.Debug(ctx, "Primary connection replaced, retrying request", log.Int("attempt", attempt))
		}
	}
}

// errRetryableOnNewConn reports whether request failed because connection
// died before the request was processed by the server (request was not sent,
// or sent but not acknowledged), so it is safe to retry the request on a new
// connection.
func errRetryableOnNewConn(err error) bool {
	return errors.Is(err, pool.ErrConnDead) || errors.Is(err, rpc.ErrEngineClosed)
}
