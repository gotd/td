package helpers

import (
	"context"
	"time"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/telegram"
)

// FloodWait sleeps required duration and true if err is FLOOD_WAIT
// or false and context or original error otherwise.
func FloodWait(ctx context.Context, err error) (bool, error) {
	if d, ok := telegram.AsFloodWait(err); ok {
		timer := clock.System.Timer(d + 1*time.Second)
		defer clock.StopTimer(timer)

		select {
		case <-timer.C():
			return true, err
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	return false, err
}
