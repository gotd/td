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

// WaitScheduler is a invoker middleware to handle FLOOD_WAIT errors from Telegram.
type WaitScheduler struct {
	prev  tg.Invoker // immutable
	clock clock.Clock
	sch   *scheduler

	tick       time.Duration
	waitLimit  int
	retryLimit int
}

// NewWaitScheduler creates new WaitScheduler invoker middleware.
func NewWaitScheduler(prev tg.Invoker) *WaitScheduler {
	return &WaitScheduler{
		prev:       prev,
		clock:      clock.System,
		sch:        newScheduler(clock.System, time.Second),
		tick:       1 * time.Millisecond,
		waitLimit:  60,
		retryLimit: 5,
	}
}

// WithClock sets clock to use.
func (w *WaitScheduler) WithClock(c clock.Clock) *WaitScheduler {
	w.clock = c
	return w
}

// WithWaitLimit sets wait limit to use.
func (w *WaitScheduler) WithWaitLimit(waitLimit int) *WaitScheduler {
	if waitLimit >= 0 {
		w.waitLimit = waitLimit
	}
	return w
}

// WithRetryLimit sets retry limit to use.
func (w *WaitScheduler) WithRetryLimit(retryLimit int) *WaitScheduler {
	if retryLimit >= 0 {
		w.retryLimit = retryLimit
	}
	return w
}

// WithTick sets gather tick for WaitScheduler.
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
