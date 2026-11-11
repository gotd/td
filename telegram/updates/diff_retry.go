package updates

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/tgerr"
)

const (
	// diffRetryMin is the first delay before retrying a failed difference.
	diffRetryMin = time.Second
	// diffRetryMax caps the doubling delay; diffRetryJitter is added at the
	// cap, so a long outage retries every 60–80 s like TDLib
	// (UpdatesManager::on_failed_get_difference).
	diffRetryMax    = 60 * time.Second
	diffRetryJitter = 20 * time.Second
)

// diffRetry schedules retries of a failed difference with backoff. Telegram
// Desktop (failDifferenceStartTimerFor) and TDLib retry a failed
// getDifference/getChannelDifference with doubling delays; without it a
// failed difference waited for the next update with a gap or the 15-minute
// idle timer.
//
// It is used only from the goroutine that owns the state.
type diffRetry struct {
	clock clock.Clock
	delay time.Duration
	timer clock.Timer
	// C fires when a retry is due; nil when none is scheduled.
	C <-chan time.Time
}

func newDiffRetry(c clock.Clock) *diffRetry {
	if c == nil {
		c = clock.System
	}
	return &diffRetry{clock: c}
}

// schedule schedules a retry after err and returns the delay. A FLOOD_WAIT is
// waited out at least.
func (r *diffRetry) schedule(err error) time.Duration {
	switch {
	case r.delay == 0:
		r.delay = diffRetryMin
	case r.delay < diffRetryMax:
		r.delay = min(r.delay*2, diffRetryMax)
	}
	d := r.delay
	if d >= diffRetryMax {
		d += time.Duration(rand.Int64N(int64(diffRetryJitter)))
	}
	if wait, ok := tgerr.AsFloodWait(err); ok && wait > d {
		d = wait
	}
	r.stopTimer()
	r.timer = r.clock.Timer(d)
	r.C = r.timer.C()
	return d
}

// fired must be called when C fired.
func (r *diffRetry) fired() {
	r.timer = nil
	r.C = nil
}

// reset clears the backoff after a successful difference.
func (r *diffRetry) reset() {
	r.delay = 0
	r.stopTimer()
}

func (r *diffRetry) stopTimer() {
	if r.timer != nil {
		clock.StopTimer(r.timer)
		r.timer = nil
	}
	r.C = nil
}

// retryableDiffError reports whether a failed difference should be retried
// later: not on cancellation, on errors that stop the manager, or on a lost
// authorization, which no retry fixes.
func retryableDiffError(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) && ctx.Err() != nil {
		return false
	}
	return !isFatalError(err)
}

// isPersistentTimestampReset reports whether the server rejected the local
// state of the common sequence: TDLib resets it with updates.getState
// (PERSISTENT_TIMESTAMP_INVALID, and EMPTY for a state it does not know).
func isPersistentTimestampReset(err error) bool {
	return tgerr.Is(err, "PERSISTENT_TIMESTAMP_INVALID", "PERSISTENT_TIMESTAMP_EMPTY")
}
