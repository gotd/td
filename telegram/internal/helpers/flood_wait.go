package helpers

import (
	"context"
	"time"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/tgerr"
)

// ErrFloodWait is error type of "FLOOD_WAIT" error.
const ErrFloodWait = "FLOOD_WAIT"

// AsFloodWait returns wait duration and true boolean if err is
// the "FLOOD_WAIT" error.
//
// Client should wait for that duration before issuing new requests with
// same method.
func AsFloodWait(err error) (d time.Duration, ok bool) {
	if rpcErr, ok := tgerr.AsType(err, ErrFloodWait); ok {
		return time.Second * time.Duration(rpcErr.Argument), true
	}
	return 0, false
}

// FloodWait sleeps required duration and true if err is FLOOD_WAIT
// or false and context or original error otherwise.
func FloodWait(ctx context.Context, err error) (bool, error) {
	if d, ok := AsFloodWait(err); ok {
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
