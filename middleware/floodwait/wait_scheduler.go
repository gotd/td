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

// WaitScheduler is a tg.Invoker that handles flood wait errors on underlying invoker.
//
// This implementation uses a request scheduler and is more suitable for long-running
// programs with high level of concurrency and parallelism.
//
// You should use WaitScheduler if unsure which waiter implementation to use.
//
// See WaitTimer for a simple timer-based implementation.
type WaitScheduler struct {
	next  tg.Invoker // immutable
	clock clock.Clock
	sch   *scheduler

	tick       time.Duration
	maxWait    time.Duration
	maxRetries int
}

// NewWaitScheduler returns a new invoker that waits on the flood wait errors.
func NewWaitScheduler(invoker tg.Invoker) *WaitScheduler {
	return &WaitScheduler{
		next:       invoker,
		clock:      clock.System,
		sch:        newScheduler(clock.System, time.Second),
		tick:       time.Millisecond,
		maxWait:    60 * time.Second,
		maxRetries: 5,
	}
}

// clone returns a copy of the WaitScheduler.
func (w *WaitScheduler) clone() *WaitScheduler {
	return &WaitScheduler{
		next:       w.next,
		clock:      w.clock,
		sch:        w.sch,
		tick:       w.tick,
		maxWait:    w.maxWait,
		maxRetries: w.maxRetries,
	}
}

// WithClock sets clock to use. Default is to use system clock.
func (w *WaitScheduler) WithClock(c clock.Clock) *WaitScheduler {
	w = w.clone()
	w.clock = c
	return w
}

// WithMaxWait limits wait time per attempt. WaitScheduler will return an error if flood wait
// time exceeds that limit. Default is to wait at most a minute.
//
// To limit total wait time use a context.Context with timeout or deadline set.
func (w *WaitScheduler) WithMaxWait(m time.Duration) *WaitScheduler {
	w = w.clone()
	w.maxWait = m
	return w
}

// WithMaxRetries sets max number of retries before giving up. Default is to retry at most 5 times.
func (w *WaitScheduler) WithMaxRetries(m int) *WaitScheduler {
	w = w.clone()
	w.maxRetries = m
	return w
}

// WithTick sets gather tick interval for WaitScheduler. Default is 1ms.
func (w *WaitScheduler) WithTick(t time.Duration) *WaitScheduler {
	w = w.clone()
	if t <= 0 {
		t = time.Nanosecond
	}
	w.tick = t
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
	err := w.next.InvokeRaw(s.request.ctx, s.request.input, s.request.output)

	floodWait, ok := tgerr.AsType(err, ErrFloodWait)
	if !ok {
		w.sch.nice(s.request.key)
		return true, err
	}

	s.request.retry++

	if max := w.maxRetries; max != 0 && s.request.retry > max {
		return true, xerrors.Errorf("flood wait retry limit exceeded (%d > %d): %w", s.request.retry, max, err)
	}

	arg := floodWait.Argument
	if arg <= 0 {
		arg = 1
	}
	d := time.Duration(arg) * time.Second

	if max := w.maxWait; max != 0 && d > max {
		return true, xerrors.Errorf("flood wait argument is too big (%v > %v): %w", d, max, err)
	}

	w.sch.flood(s.request, d)
	return false, nil
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
