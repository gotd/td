package invokers

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type request struct {
	timer       *time.Ticker
	lastTimeout time.Duration
	mux         sync.Mutex
}

func (r *request) updateTimer(arg time.Duration) {
	r.mux.Lock()
	if r.lastTimeout < arg {
		r.lastTimeout = arg
	}
	r.timer.Reset(arg)
	r.mux.Unlock()
}

func (r *request) decrementTimer(arg time.Duration) bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.lastTimeout-arg > 0 {
		r.lastTimeout -= arg
		r.timer.Reset(r.lastTimeout)
		return true
	}

	return false
}

// Waiter is a invoker middleware to handle FLOOD_WAIT errors from Telegram.
type Waiter struct {
	prev       tg.Invoker // immutable
	waiters    map[uint32]*request
	waitersMux sync.RWMutex

	retryLimit int           // immutable
	waitLimit  time.Duration // immutable
}

// NewWaiter creates new Waiter invoker middleware.
func NewWaiter(prev tg.Invoker) *Waiter {
	return &Waiter{
		prev:       prev,
		waiters:    map[uint32]*request{},
		retryLimit: 5,
		waitLimit:  60 * time.Second,
	}
}

// WithRetryLimit sets retry limit for Waiter.
func (w *Waiter) WithRetryLimit(retryLimit int) *Waiter {
	w.retryLimit = retryLimit
	return w
}

// WithWaitLimit sets wait limit for Waiter.
func (w *Waiter) WithWaitLimit(waitLimit time.Duration) *Waiter {
	w.waitLimit = waitLimit
	return w
}

// Object is a abstraction for Telegram API object with TypeID.
type Object interface {
	TypeID() uint32
}

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = "FLOOD_WAIT"

// InvokeRaw implements tg.Invoker.
func (w *Waiter) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	obj, ok := input.(Object)
	if !ok {
		return w.prev.InvokeRaw(ctx, input, output)
	}
	typeID := obj.TypeID()

	retries := 0
	for {
		retries++
		if retries > w.retryLimit {
			return xerrors.Errorf("retry limit exceeded (%d > %d)", retries, w.retryLimit)
		}

		ok, err := w.try(ctx, input, output, typeID)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		return nil
	}
}

func getTimeout(rpcErr *tgerr.Error) time.Duration {
	return time.Duration(rpcErr.Argument+1) * time.Second
}

// try tries to invoke, waits if got FLOOD_WAIT error and tries to invoke again.
// Returns:
//
// 	(true, nil) — Successful invoke.
// 	(false, nil) — Got FLOOD_WAIT.
// 	(false, err) — Got another RPC error.
//
func (w *Waiter) try(ctx context.Context, input bin.Encoder, output bin.Decoder, typeID uint32) (bool, error) {
	// Check if timer already exist.
	w.waitersMux.RLock()
	req, ok := w.waiters[typeID]
	w.waitersMux.RUnlock()

	if !ok {
		// If not, try to invoke first time.
		return w.sendNew(ctx, input, output, typeID)
	}

	select {
	case <-req.timer.C:
		// If timer already exist, wait for next tick and try to invoke.
		err := w.prev.InvokeRaw(ctx, input, output)
		rpcErr, ok := tgerr.AsType(err, ErrFloodWait)

		// If result is not a FLOOD_WAIT, decrease timeout and return result.
		if !ok {
			if !req.decrementTimer(time.Second) {
				// If timeout too small, delete timer.
				w.waitersMux.Lock()
				delete(w.waiters, typeID)
				w.waitersMux.Unlock()
			}
			return err == nil, err
		}

		// Otherwise we increase timeout.
		timeout := getTimeout(rpcErr)
		if timeout > w.waitLimit {
			return false, xerrors.Errorf("wait timeout is too big (%v > %v)", timeout, w.waitLimit)
		}
		req.updateTimer(timeout)
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (w *Waiter) sendNew(ctx context.Context, input bin.Encoder, output bin.Decoder, typeID uint32) (bool, error) {
	err := w.prev.InvokeRaw(ctx, input, output)
	rpcErr, ok := tgerr.AsType(err, ErrFloodWait)
	if !ok {
		return err == nil, err
	}

	// If got FLOOD_WAIT, try to create or get existing timer.
	timeout := getTimeout(rpcErr)
	w.waitersMux.Lock()
	req, ok := w.waiters[typeID]
	if !ok {
		req = &request{
			timer:       time.NewTicker(timeout),
			lastTimeout: timeout,
		}
		w.waiters[typeID] = req
		w.waitersMux.Unlock()
		return false, nil
	}
	w.waitersMux.Unlock()

	// Increase timeout.
	req.updateTimer(timeout)
	return false, nil
}
