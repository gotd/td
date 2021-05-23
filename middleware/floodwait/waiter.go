package floodwait

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Waiter is a tg.Invoker that handles flood wait errors on underlying invoker.
//
// This implementation is more suitable for one-off tasks and programs with low level
// of concurrency and parallelism.
type Waiter struct {
	next  tg.Invoker
	clock clock.Clock

	maxRetries uint
	maxWait    time.Duration
}

// NewWaiter returns a new invoker that waits on the flood wait errors.
func NewWaiter(invoker tg.Invoker) *Waiter {
	return &Waiter{
		next:  invoker,
		clock: clock.System,
	}
}

// clone returns a copy of the Waiter.
func (w *Waiter) clone() *Waiter {
	return &Waiter{
		next:       w.next,
		clock:      w.clock,
		maxWait:    w.maxWait,
		maxRetries: w.maxRetries,
	}
}

// WithClock sets clock to use. Default is to use system clock.
func (w *Waiter) WithClock(c clock.Clock) *Waiter {
	w = w.clone()
	w.clock = c
	return w
}

// WithMaxRetries sets max number of retries before giving up. Default is to keep retrying
// on flood wait errors indefinitely.
func (w *Waiter) WithMaxRetries(m uint) *Waiter {
	w = w.clone()
	w.maxRetries = m
	return w
}

// WithMaxWait limits wait time per attempt. Waiter will return an error if flood wait
// time exceeds that limit. Default is to wait without time limit.
//
// To limit total wait time use a context.Context with timeout or deadline set.
func (w *Waiter) WithMaxWait(m time.Duration) *Waiter {
	w = w.clone()
	w.maxWait = m
	return w
}

// InvokeRaw implements tg.Invoker.
func (w *Waiter) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	var t clock.Timer

	var retries uint
	for {
		err := w.next.InvokeRaw(ctx, input, output)
		if err == nil {
			return nil
		}

		floodWait, ok := tgerr.AsType(err, tgerr.ErrFloodWait)
		if !ok {
			return err
		}

		retries++

		if max := w.maxRetries; max != 0 && retries > max {
			return xerrors.Errorf("flood wait retry limit exceeded (%d > %d): %w", retries, max, err)
		}

		arg := floodWait.Argument
		if arg <= 0 {
			arg = 1
		}
		d := time.Duration(arg) * time.Second

		if max := w.maxWait; max != 0 && d > max {
			return xerrors.Errorf("flood wait argument is too big (%v > %v): %w", d, max, err)
		}

		if t == nil {
			t = w.clock.Timer(d)
		} else {
			clock.StopTimer(t)
			t.Reset(d)
		}
		select {
		case <-t.C():
			continue
		case <-ctx.Done():
			clock.StopTimer(t)
			return ctx.Err()
		}
	}
}
