package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
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

// measuredInvoker records gotd.client.invoke.duration of requests through
// the client invoker, with middlewares: Client.Invoke and Client.API.
type measuredInvoker struct {
	client *Client
	next   tg.Invoker
}

func (i measuredInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	c := i.client
	start := c.now()
	err := i.next.Invoke(ctx, input, output)
	m := c.meters()
	m.invokeDuration.Record(ctx, c.now().Sub(start).Seconds(), m.invokeResult(ctx, err)...)
	return err
}

// invokeDirect directly invokes RPC method, automatically handling datacenter redirects.
func (c *Client) invokeDirect(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	err := c.invokeConn(ctx, input, output)
	if err != nil {
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
				c.meters().fileMigrations.Add(ctx, 1, dcAttrs(targetDC,
					attribute.String("source", "primary"),
				))
				return c.invokeSub(ctx, targetDC, input, output)
			}

			// Otherwise we should change primary DC.
			logger.Info(ctx, "Migrating to target DC")
			err = c.invokeMigrate(ctx, targetDC, input, output)
			c.observeAuth(input, output, err)
			return err
		}

		c.observeAuth(input, output, err)
		return err
	}

	// Only primary connection results reveal primary session authorization:
	// sub-connection (FILE_MIGRATE) results are not observed.
	c.observeAuth(input, output, nil)
	return nil
}

// invokeConn directly invokes RPC call on primary connection without any
// additional handling.
//
// If the connection dies before the request is processed by the server,
// invokeConn waits until the reconnection loop replaces the connection and
// retries the request on it, see https://github.com/gotd/td/issues/1030.
func (c *Client) invokeConn(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	var timeout time.Duration
	if c.requestTimeout != nil {
		timeout = c.requestTimeout(input)
	}
	attempt := 0
	resent := false
	for {
		c.connMux.Lock()
		conn := c.conn
		connChanged := c.connChanged
		c.connMux.Unlock()

		reqCtx, cancel := ctx, context.CancelCauseFunc(func(error) {})
		if timeout > 0 {
			reqCtx, cancel = c.withResponseDeadline(ctx, conn, timeout)
		}
		err := conn.Invoke(reqCtx, input, output)
		timedOut := err != nil && ctx.Err() == nil && errors.Is(context.Cause(reqCtx), errNoResponse)
		cancel(nil)
		if timedOut {
			// Response deadline passed, but caller still waits: connection is
			// suspected to be dead.
			method := methodName(input)
			if resent {
				c.meters().requestTimeout.Add(ctx, 1, metric.WithAttributes(
					attribute.String("result", "failed"),
					attribute.String("method", method),
				))
				c.log.Warn(ctx, "No response in time after resend",
					log.Duration("timeout", timeout),
					log.String("method", method),
				)
				return errors.Errorf("no response in %s: %w", timeout, context.DeadlineExceeded)
			}
			resent = true
			c.meters().requestTimeout.Add(ctx, 1, metric.WithAttributes(
				attribute.String("result", "resent"),
				attribute.String("method", method),
			))
			c.log.Warn(ctx, "No response in time, reconnecting to resend request",
				log.Duration("timeout", timeout),
				log.String("method", method),
			)
			if err := c.restartSuspect(ctx, connChanged); err != nil {
				return errors.Wrap(err, "reconnect")
			}
			continue
		}
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

// methodName returns TL name of the innermost request, e.g. "messages.getHistory".
func methodName(input bin.Encoder) string {
	if named, ok := unwrapQuery(input).(interface{ TypeName() string }); ok {
		return named.TypeName()
	}
	return "unknown"
}

// errRetryableOnNewConn reports whether request failed because connection
// died before the request was processed by the server (request was not sent,
// or sent but not acknowledged), so it is safe to retry the request on a new
// connection.
func errRetryableOnNewConn(err error) bool {
	return errors.Is(err, pool.ErrConnDead) || errors.Is(err, rpc.ErrEngineClosed)
}

// restartSuspect restarts primary connection, unless it was already replaced
// (connChanged is closed), and waits until new connection is ready.
func (c *Client) restartSuspect(ctx context.Context, connChanged <-chan struct{}) error {
	// Another request may be restarting a connection that never gets ready:
	// do not wait for it past own context.
	select {
	case c.restartSem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-c.restartSem }()

	select {
	case <-connChanged:
		// Already reconnected, e.g. by another timed out request.
		return nil
	default:
	}
	return c.ensureRestart(ctx, restartTimeout)
}

// errNoResponse is the cause of request context cancellation when response
// deadline passes.
var errNoResponse = errors.New("no response in time")

// withResponseDeadline returns context that is canceled with errNoResponse
// once timeout passes after conn became ready to send requests.
//
// Time spent waiting for the connection (dial, key exchange, initConnection)
// does not count: a slow handshake must not be taken for a dead connection,
// or each request would restart it before it gets ready.
func (c *Client) withResponseDeadline(
	ctx context.Context, conn clientConn, timeout time.Duration,
) (context.Context, context.CancelCauseFunc) {
	reqCtx, cancel := context.WithCancelCause(ctx)

	var ready <-chan struct{}
	if r, ok := conn.(interface{ Ready() <-chan struct{} }); ok {
		ready = r.Ready()
	}
	clk := c.clock
	if clk == nil {
		clk = clock.System
	}
	go func() {
		if ready != nil {
			select {
			case <-ready:
			case <-reqCtx.Done():
				return
			}
		}
		t := clk.Timer(timeout)
		defer clock.StopTimer(t)
		select {
		case <-t.C():
			cancel(errNoResponse)
		case <-reqCtx.Done():
		}
	}()

	return reqCtx, cancel
}
