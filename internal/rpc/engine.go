package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

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
	log           *zap.Logger
	retryInterval time.Duration
	maxRetries    int

	// Canceling pending requests in ForceClose.
	reqCtx    context.Context
	reqCancel context.CancelFunc

	wg     sync.WaitGroup
	closed uint32
}

// New creates new rpc Engine.
func New(send Send, cfg Options) *Engine {
	cfg.setDefaults()

	cfg.Logger.Info("Initialized",
		zap.Duration("retry_interval", cfg.RetryInterval),
		zap.Int("max_retries", cfg.MaxRetries),
	)

	reqCtx, reqCancel := context.WithCancel(context.Background())
	return &Engine{
		rpc: map[int64]func(*bin.Buffer, error) error{},
		ack: map[int64]chan struct{}{},

		send: send,
		drop: cfg.DropHandler,

		log:           cfg.Logger,
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
	if e.isClosed() {
		return ErrEngineClosed
	}

	e.wg.Add(1)
	defer e.wg.Done()

	retryCtx, retryClose := context.WithCancel(ctx)
	defer retryClose()

	log := e.log.With(zap.Int64("msg_id", req.MsgID))
	log.Debug("Do called")

	done := make(chan struct{})

	var (
		// Handler result.
		resultErr error
		// Needed to prevent multiple handler calls.
		handlerCalled uint32
	)

	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		log.Debug("Handler called")

		if ok := atomic.CompareAndSwapUint32(&handlerCalled, 0, 1); !ok {
			log.Warn("Handler already called")

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

	select {
	case <-ctx.Done():
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
			log.Info("Failed to drop request", zap.Error(err))
			return ctx.Err()
		}

		log.Debug("Request dropped")
		return ctx.Err()
	case <-e.reqCtx.Done():
		return errors.Wrap(e.reqCtx.Err(), "engine forcibly closed")
	case <-done:
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
		log     = e.log.Named("retry").With(zap.Int64("msg_id", req.MsgID))
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
				return errors.Wrap(e.reqCtx.Err(), "engine forcibly closed")
			case <-ackChan:
				log.Debug("Acknowledged")
				return nil
			case <-timer.C():
				timer.Reset(e.retryInterval)

				log.Debug("Acknowledge timed out, performing retry")
				if err := e.send(ctx, req.MsgID, req.SeqNo, req.Input); err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}

					log.Error("Retry failed", zap.Error(err))
					return err
				}

				retries++
				if retries >= e.maxRetries {
					log.Error("Retry limit reached", zap.Int64("msg_id", req.MsgID))
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
	e.mux.Unlock()
	if !ok {
		e.log.Warn("rpc callback not set", zap.Int64("msg_id", msgID))
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
		e.log.Warn("rpc callback not set", zap.Int64("msg_id", msgID))
		return
	}

	// Callback with rpcError always return nil.
	_ = fn(nil, rpcErr)
}

func (e *Engine) isClosed() bool {
	return atomic.LoadUint32(&e.closed) == 1
}

// Close gracefully closes the engine.
// All pending requests will be awaited.
// All Do method calls of closed engine will return ErrEngineClosed error.
func (e *Engine) Close() {
	atomic.StoreUint32(&e.closed, 1)
	e.log.Info("Close called")
	e.wg.Wait()
}

// ForceClose forcibly closes the engine.
// All pending requests will be canceled.
// All Do method calls of closed engine will return ErrEngineClosed error.
func (e *Engine) ForceClose() {
	e.reqCancel()
	e.Close()
}
