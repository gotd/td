package invokers

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// WaitScheduler is a tg.Invoker that handles flood wait errors on underlying invoker.
//
// This implementation uses a request scheduler and is more suitable for long-running
// programs with high level of concurrency and parallelism.
//
// You should use WaitScheduler if unsure which waiter implementation to use.
//
// See Waiter for a simple sleep-based implementation.
type WaitScheduler struct {
	prev  tg.Invoker // immutable
	clock clock.Clock
	sch   *scheduler

	tick       time.Duration
	waitLimit  int
	retryLimit int
}

// NewWaitScheduler returns a new invoker that waits on the flood wait errors.
func NewWaitScheduler(prev tg.Invoker) *WaitScheduler {
	return &WaitScheduler{
		prev:       prev,
		clock:      clock.System,
		sch:        newScheduler(clock.System, time.Second),
		tick:       time.Millisecond,
		waitLimit:  60,
		retryLimit: 5,
	}
}

// WithClock sets clock to use. Default is to use system clock.
func (w *WaitScheduler) WithClock(c clock.Clock) *WaitScheduler {
	w.clock = c
	return w
}

// WithWaitLimit limits wait time per attempt. WaitScheduler will return an error if flood wait
// time exceeds that limit. Default is to wait at most a minute.
//
// To limit total wait time use a context.Context with timeout or deadline set.
func (w *WaitScheduler) WithWaitLimit(waitLimit int) *WaitScheduler {
	if waitLimit >= 0 {
		w.waitLimit = waitLimit
	}
	return w
}

// WithRetryLimit sets max number of retries before giving up. Default is to retry at most 5 times.
func (w *WaitScheduler) WithRetryLimit(retryLimit int) *WaitScheduler {
	if retryLimit >= 0 {
		w.retryLimit = retryLimit
	}
	return w
}

// WithTick sets gather tick interval for WaitScheduler. Default is 1ms.
func (w *WaitScheduler) WithTick(tick time.Duration) *WaitScheduler {
	if tick > 0 {
		w.tick = tick
	}
	return w
}

// Run runs send loop.
func (w *WaitScheduler) Run(ctx context.Context) error {
	ticker := w.clock.Ticker(w.tick)
	defer ticker.Stop()

	var requests []scheduled
	for {
		select {
		case <-ticker.C():
			requests = w.sch.gather(requests[:0])
			if len(requests) < 1 {
				continue
			}

			for _, s := range requests {
				ret, err := w.send(s)
				if ret {
					select {
					case s.request.result <- err:
					default:
					}
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *WaitScheduler) send(s scheduled) (bool, error) {
	err := w.prev.InvokeRaw(s.request.ctx, s.request.input, s.request.output)

	floodWait, ok := tgerr.AsType(err, ErrFloodWait)
	switch {
	case !ok:
		w.sch.nice(s.request.key)
		return true, err
	case floodWait.Argument >= w.waitLimit:
		return true, xerrors.Errorf("FLOOD_WAIT argument is too big (%d >= %d)", floodWait.Argument, w.waitLimit)
	case s.request.retry >= w.retryLimit:
		return true, xerrors.Errorf("retry limit exceeded (%d >= %d)", s.request.retry, w.retryLimit)
	}

	s.request.retry++
	w.sch.flood(s.request, time.Duration(floodWait.Argument)*time.Second)
	return false, nil
}

// Object is a abstraction for Telegram API object with TypeID.
type Object interface {
	TypeID() uint32
}

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = "FLOOD_WAIT"

// InvokeRaw implements tg.Invoker.
func (w *WaitScheduler) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	select {
	case err := <-w.sch.new(ctx, input, output):
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
