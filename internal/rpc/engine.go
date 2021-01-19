package rpc

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

// Engine handles RPC requests.
type Engine struct {
	send Send

	mux sync.Mutex
	rpc map[int64]func(*bin.Buffer, error) error
	ack map[int64]chan struct{}

	clock         clock.Clock
	log           *zap.Logger
	retryInterval time.Duration
	maxRetries    int

	wg     sync.WaitGroup
	closed uint32
}

// Send is a function that sends requests to the server.
type Send func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error

// NopSend does nothing.
func NopSend(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error { return nil }

// New creates new rpc Engine.
func New(send Send, cfg Options) *Engine {
	cfg.setDefaults()

	cfg.Logger.Info("Initialized",
		zap.Duration("retry_interval", cfg.RetryInterval),
		zap.Int("max_retries", cfg.MaxRetries),
	)

	return &Engine{
		rpc: map[int64]func(*bin.Buffer, error) error{},
		ack: map[int64]chan struct{}{},

		send: send,

		log:           cfg.Logger,
		maxRetries:    cfg.MaxRetries,
		retryInterval: cfg.RetryInterval,
		clock:         cfg.Clock,
	}
}

// Request represents client RPC request.
type Request struct {
	ID       int64
	Sequence int32
	Input    bin.Encoder
	Output   bin.Decoder
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

	log := e.log.With(zap.Int64("msg_id", req.ID))
	log.Debug("Do called")

	done := make(chan struct{})

	var (
		// Handler result.
		resultErr error
		// Needed to prevent multiple handler calls.
		handlerCalls uint32
	)

	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		log.Debug("Handler called")

		if calls := atomic.AddUint32(&handlerCalls, 1); calls > 1 {
			log.Warn("Handler already called")

			return xerrors.Errorf("handler already called")
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
	e.rpc[req.ID] = handler
	e.mux.Unlock()

	defer func() {
		// Ensuring that callback can't be called after function return.
		e.mux.Lock()
		delete(e.rpc, req.ID)
		e.mux.Unlock()
	}()

	// Start retrying.
	if err := e.retryUntilAck(retryCtx, req); err != nil && !errors.Is(err, context.Canceled) {
		// If the retryCtx was canceled, then one of two things happened:
		//   1. User canceled the original context.
		//   2. The RPC result came and callback canceled retryCtx.
		//
		// If this is not an context.Canceled error, most likely we did not receive ack
		// and exceeded the limit of attempts to send a request,
		// or could not write data to the connection, so we return an error.
		return xerrors.Errorf("retryUntilAck: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return resultErr
	}
}

// retryUntilAck resends the request to the server until request is
// acknowledged.
//
// Returns nil if acknowledge was received or error otherwise.
func (e *Engine) retryUntilAck(ctx context.Context, req Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		ackChan = e.waitAck(req.ID)
		retries = 0
		log     = e.log.Named("retry").With(zap.Int64("msg_id", req.ID))
	)

	defer e.removeAck(req.ID)

	// Encoding request.
	if err := e.send(ctx, req.ID, req.Sequence, req.Input); err != nil {
		return xerrors.Errorf("send: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ackChan:
			log.Debug("Acknowledged")
			return nil
		case <-e.clock.After(e.retryInterval):
			log.Debug("Acknowledge timed out, performing retry")
			if err := e.send(ctx, req.ID, req.Sequence, req.Input); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}

				log.Error("Retry failed", zap.Error(err))
				return err
			}

			retries++
			if retries >= e.maxRetries {
				log.Error("Retry limit reached", zap.Int64("msg_id", req.ID))
				return &RetryLimitReachedErr{
					Retries: retries,
				}
			}
		}
	}
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

// Close closes the engine.
// All Do method calls of closed engine will return ErrEngineClosed error.
func (e *Engine) Close() {
	atomic.StoreUint32(&e.closed, 1)
	e.log.Info("Close called")
	e.wg.Wait()
}
