package telegram

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
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

// invokeConn directly invokes RPC call on primary connection without any
// additional handling.
func (c *Client) invokeConn(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c.connMux.Lock()
	conn := c.conn
	c.connMux.Unlock()

	return conn.Invoke(ctx, input, output)
}
