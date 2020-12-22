// Package rpc implements rpc engine.
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
)

// Config of rpc engine.
type Config struct {
	RetryInterval time.Duration
	MaxRetries    int
}

// Engine handles RPC requests.
type Engine struct {
	rpc    map[int64]func(*bin.Buffer, error) error
	rpcMux sync.RWMutex

	ack    map[int64]func()
	ackMux sync.RWMutex

	client Sender
	log    *zap.Logger
	cfg    Config
}

// NewEngine creates new rpc engine.
func NewEngine(client Sender, log *zap.Logger, cfg Config) *Engine {
	return &Engine{
		rpc: map[int64]func(*bin.Buffer, error) error{},
		ack: map[int64]func(){},

		client: client,
		log:    log,
		cfg:    cfg,
	}
}

// Request represents client RPC request.
type Request struct {
	ID       int64
	Sequence int32
	Input    bin.Encoder
	Output   bin.Decoder
}

// DoRequest performs RPC request.
func (e *Engine) DoRequest(ctx context.Context, req Request) error {
	retryCtx, retryClose := context.WithCancel(ctx)
	defer retryClose()

	log := e.log.With(zap.Int64("msg_id", req.ID))

	var (
		// Shows that we received a response.
		doneCh = make(chan struct{})
		// Handler result.
		resultErr error
		// Needed to prevent multiple handler calls.
		handlerCalls uint32
	)

	handler := func(rpcBuff *bin.Buffer, rpcErr error) error {
		log.Info("handler called")

		atomic.AddUint32(&handlerCalls, 1)
		if atomic.LoadUint32(&handlerCalls) > 1 {
			log.Warn("handler already called")

			return xerrors.Errorf("handler already called")
		}

		defer retryClose()
		defer close(doneCh)

		if rpcErr != nil {
			resultErr = rpcErr
			return nil
		}

		resultErr = req.Output.Decode(rpcBuff)
		return resultErr
	}

	// Setting callback that will be called if message is received.
	e.rpcMux.Lock()
	e.rpc[req.ID] = handler
	e.rpcMux.Unlock()

	defer func() {
		// Ensuring that callback can't be called after function return.
		e.rpcMux.Lock()
		delete(e.rpc, req.ID)
		e.rpcMux.Unlock()
	}()

	// Encoding request. Note that callback is already set.
	if err := e.client.Send(ctx, req.ID, req.Sequence, req.Input); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	// Start retrying.
	if err := e.rpcRetryUntilAck(retryCtx, req); err != nil {
		// If the retryCtx was canceled, then one of two things happened:
		// 1. User canceled the original context.
		// 2. The RPC result came and callback canceled retryCtx.
		//
		// If this is not an context.Canceled error, most likely we did not receive ACK
		// and exceeded the limit of attempts to send a request,
		// or could not write data to the connection, so we return an error.
		if !errors.Is(err, context.Canceled) {
			return xerrors.Errorf("retryUntilAck: %w", err)
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return resultErr
	}
}

// rpcRetryUntilAck resends the request to the server until ACK is received
// or context canceled.
//
// Returns nil if ACK was received, otherwise return error.
func (e *Engine) rpcRetryUntilAck(ctx context.Context, req Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ackChan := make(chan error)
	go func() { ackChan <- e.waitACK(ctx, req.ID) }()

	retries := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ackErr := <-ackChan:
			if ackErr != nil {
				return xerrors.Errorf("wait ack: %w", ackErr)
			}

			return nil
			// TODO(ccln): use clock.
		case <-time.After(e.cfg.RetryInterval):
			e.log.Info("RPC Retrying", zap.Int64("msg_id", req.ID))
			if err := e.client.Send(ctx, req.ID, req.Sequence, req.Input); err != nil {
				e.log.Error("Retry attempt failed", zap.Error(err))
				return err
			}

			retries++
			if retries >= e.cfg.MaxRetries {
				e.log.Error("Retry limit reached", zap.Int64("request_id", req.ID))
				return xerrors.New("retry limit reached")
			}
		}
	}
}

// NotifyResult notifies engine about received RPC response.
func (e *Engine) NotifyResult(msgID int64, b *bin.Buffer) error {
	e.rpcMux.Lock()
	fn, ok := e.rpc[msgID]
	e.rpcMux.Unlock()
	if !ok {
		e.log.Warn("rpc callback not set", zap.Int64("msg_id", msgID))
		return nil
	}

	return fn(b, nil)
}

// NotifyError notifies engine about received RPC error.
func (e *Engine) NotifyError(msgID int64, rpcErr error) {
	e.rpcMux.Lock()
	fn, ok := e.rpc[msgID]
	e.rpcMux.Unlock()
	if !ok {
		e.log.Warn("rpc callback not set", zap.Int64("msg_id", msgID))
		return
	}

	// Callback with rpcError always return nil.
	_ = fn(nil, rpcErr)
}
