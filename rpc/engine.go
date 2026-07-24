package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

// Engine handles RPC requests.
type Engine struct {
	send Send
	drop DropHandler

	mux sync.Mutex
	rpc map[int64]func(*bin.Buffer, error) error
	ack map[int64]chan struct{}

	clock         clock.Clock
	log           log.Helper
	retryInterval time.Duration
	maxRetries    int

	// Canceling pending requests in ForceClose.
	reqCtx    context.Context
	reqCancel context.CancelCauseFunc

	wg sync.WaitGroup
	// closed is guarded by mux: it must be set and observed atomically with
	// respect to wg.Add in Do, otherwise Close may observe a zero WaitGroup
	// counter and return from Wait concurrently with a new Add.
	closed bool
}

// New creates new rpc Engine.
func New(send Send, cfg Options) *Engine {
	cfg.setDefaults()

	logger := log.For(cfg.Logger)
	logger.Info(context.Background(), "Initialized",
		log.Duration("retry_interval", cfg.RetryInterval),
		log.Int("max_retries", cfg.MaxRetries),
	)

	reqCtx, reqCancel := context.WithCancelCause(context.Background())
	return &Engine{
		rpc: map[int64]func(*bin.Buffer, error) error{},
		ack: map[int64]chan struct{}{},

		send: send,
		drop: cfg.DropHandler,

		log:           logger,
		maxRetries:    cfg.MaxRetries,
		retryInterval: cfg.RetryInterval,
		clock:         cfg.Clock,

		reqCtx:    reqCtx,
		reqCancel: reqCancel,
	}
}

// Request represents client RPC request.
type Request struct {
	MsgID  int64
	SeqNo  int32
	Input  bin.Encoder
	Output bin.Decoder
}

// Do sends request to server and blocks until response is received, performing
// multiple retries if needed.
func (e *Engine) Do(ctx context.Context, req Request) error {
	// Register the request under the mutex so Close cannot observe a zero
	// WaitGroup counter and return from Wait while we are about to Add.
	e.mux.Lock()
	if e.closed {
		e.mux.Unlock()
		return ErrEngineClosed
	}
	e.wg.Add(1)
	e.mux.Unlock()
	defer e.wg.Done()

	retryCtx, retryClose := context.WithCancel(ctx)
	defer retryClose()

	logger := e.log.With(log.Int64("msg_id", req.MsgID))
	logger.Debug(ctx, "Do called")

	done := make(chan struct{})

	var (
		// Handler result.
		resultErr error
		// Needed to prevent multiple handler calls.
		handlerCalled uint32
	)

	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		logger.Debug(ctx, "Handler called")

		if ok := atomic.CompareAndSwapUint32(&handlerCalled, 0, 1); !ok {
			logger.Warn(ctx, "Handler already called")

			return errors.New("handler already called")
		}

		defer retryClose()
		defer close(done)

		if rpcErr != nil {
			resultErr = rpcErr
			return nil
		}

		resultErr = req.Output.Decode(rpcBuff)
		return resultErr
	}

	// Setting callback that will be called if message is received.
	e.mux.Lock()
	e.rpc[req.MsgID] = handler
	e.mux.Unlock()

	defer func() {
		// Ensuring that callback can't be called after function return.
		e.mux.Lock()
		delete(e.rpc, req.MsgID)
		e.mux.Unlock()
	}()

	// Start retrying.
	sent, err := e.retryUntilAck(retryCtx, req)
	if err != nil && !errors.Is(err, retryCtx.Err()) {
		// If the retryCtx was canceled, then one of two things happened:
		//   1. User canceled the parent context.
		//   2. The RPC result came and callback canceled retryCtx.
		//
		// If this is not a Context’s error, most likely we did not receive ack
		// and exceeded the limit of attempts to send a request,
		// or could not write data to the connection, so we return an error.
		return errors.Wrap(err, "retryUntilAck")
	}

	logger.Debug(ctx, "Acknowledged, waiting for result", log.Bool("sent", sent))

	select {
	case <-ctx.Done():
		logger.Debug(ctx, "Context done before result", log.Bool("sent", sent), log.Error(ctx.Err()))
		if !sent {
			return ctx.Err()
		}

		// Set nop callback because server will respond with 'RpcDropAnswer' instead of expected result.
		//
		// NOTE(ccln): We can decode 'RpcDropAnswer' here but I see no reason to do this
		// because it will also come as a response to 'RPCDropAnswerRequest'.
		//
		// https://core.telegram.org/mtproto/service_messages#cancellation-of-an-rpc-query
		e.mux.Lock()
		e.rpc[req.MsgID] = func(b *bin.Buffer, e error) error { return nil }
		e.mux.Unlock()

		if err := e.drop(req); err != nil {
			logger.Info(ctx, "Failed to drop request", log.Error(err))
			return ctx.Err()
		}

		logger.Debug(ctx, "Request dropped")
		return ctx.Err()
	case <-e.reqCtx.Done():
		select {
		case <-done:
			// Result arrived concurrently with close, prefer it.
			return resultErr
		default:
		}
		// Request was acknowledged by the server, but the response was not
		// received before close: the server may have already processed it,
		// so resending is not safe. Report an error that still satisfies
		// errors.Is(err, context.Canceled) (callers should not retry
		// transparently) while also satisfying
		// errors.Is(err, ErrEngineClosedAfterAck) so callers can tell this
		// apart from a plain caller-initiated cancellation; deliberately
		// does NOT satisfy errors.Is(err, ErrEngineClosed).
		logger.Debug(ctx, "Engine closed while waiting for result")
		return &ackedCloseError{cause: e.reqCtx.Err()}
	case <-done:
		logger.Debug(ctx, "Result received", log.Error(resultErr))
		return resultErr
	}
}

// retryUntilAck resends the request to the server until request is
// acknowledged.
//
// Returns nil if acknowledge was received or error otherwise.
func (e *Engine) retryUntilAck(ctx context.Context, req Request) (sent bool, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		ackChan = e.waitAck(req.MsgID)
		retries = 0
		logger  = e.log.Named("retry").With(log.Int64("msg_id", req.MsgID))
	)

	defer e.removeAck(req.MsgID)

	// Encoding request.
	if err := e.send(ctx, req.MsgID, req.SeqNo, req.Input); err != nil {
		return false, errors.Wrap(err, "send")
	}

	loop := func() error {
		timer := e.clock.Timer(e.retryInterval)
		defer clock.StopTimer(timer)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-e.reqCtx.Done():
				select {
				case <-ackChan:
					// Acknowledge arrived concurrently with close, prefer it.
					logger.Debug(ctx, "Acknowledged")
					return nil
				default:
				}
				// Request was sent, but not yet acknowledged: per MTProto it is
				// safe to resend it on a new connection, so report ErrEngineClosed
				// (the cancellation cause) to let callers retry.
				return errors.Wrap(context.Cause(e.reqCtx), "engine forcibly closed")
			case <-ackChan:
				logger.Debug(ctx, "Acknowledged")
				return nil
			case <-timer.C():
				timer.Reset(e.retryInterval)

				logger.Debug(ctx, "Acknowledge timed out, performing retry")
				if err := e.send(ctx, req.MsgID, req.SeqNo, req.Input); err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}

					select {
					case <-ackChan:
						// The acknowledge for an earlier send arrived over the
						// still-live read path while this retry send was parked on
						// a wedged socket. The server has the request and may have
						// already executed it, so the retry's write failure says
						// nothing about the request's fate: prefer the ack, exactly
						// as the reqCtx branch above does.
						logger.Debug(ctx, "Acknowledged")
						return nil
					default:
					}

					logger.Error(ctx, "Retry failed", log.Error(err))

					if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
						// The send failed because ctx ended, not because the
						// connection did. Do normalizes this into ctx.Err(), and
						// masking it here would defeat that; return it unchanged.
						return err
					}

					// An earlier send already reached the wire, so the server may
					// hold this msg_id even though no ack arrived before the check
					// above (an ack racing in right now is indistinguishable from
					// one that never comes). Callers must not transparently resend:
					// telegram/invoke.go and pool/pool_conn.go would do so under a
					// fresh msg_id, defeating server-side deduplication and
					// duplicating the RPC. Mask the retryable sentinels the
					// underlying error may carry — notably transport.ErrWriteFailed,
					// which is genuinely safe to retry only for the very first send.
					return &retrySendError{cause: err}
				}

				retries++
				if retries >= e.maxRetries {
					logger.Error(ctx, "Retry limit reached", log.Int64("msg_id", req.MsgID))
					return &RetryLimitReachedErr{
						Retries: retries,
					}
				}
			}
		}
	}

	return true, loop()
}

// NotifyResult notifies engine about received RPC response.
func (e *Engine) NotifyResult(msgID int64, b *bin.Buffer) error {
	e.mux.Lock()
	fn, ok := e.rpc[msgID]
	pending := len(e.rpc)
	e.mux.Unlock()
	if !ok {
		// Result arrived but no caller is waiting for it: the request likely
		// already timed out, was dropped, or the connection was replaced.
		e.log.Warn(context.Background(), "rpc callback not set (result for unknown/expired request)",
			log.Int64("msg_id", msgID),
			log.Int("pending", pending),
		)
		return nil
	}

	return fn(b, nil)
}

// NotifyError notifies engine about received RPC error.
func (e *Engine) NotifyError(msgID int64, rpcErr error) {
	e.mux.Lock()
	fn, ok := e.rpc[msgID]
	e.mux.Unlock()
	if !ok {
		e.log.Warn(context.Background(), "rpc callback not set", log.Int64("msg_id", msgID))
		return
	}

	// Callback with rpcError always return nil.
	_ = fn(nil, rpcErr)
}

// Close gracefully closes the engine.
// All pending requests will be awaited.
// All Do method calls of closed engine will return ErrEngineClosed error.
func (e *Engine) Close() {
	e.mux.Lock()
	e.closed = true
	e.mux.Unlock()
	e.log.Info(context.Background(), "Close called")
	e.wg.Wait()
}

// ForceClose forcibly closes the engine.
// All pending requests will be canceled.
// All Do method calls of closed engine will return ErrEngineClosed error.
func (e *Engine) ForceClose() {
	e.reqCancel(ErrEngineClosed)
	e.Close()
}
