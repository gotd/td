package helpers

import (
	"context"
	"time"

	"github.com/gotd/td/mtproto"
)

// FloodWait sleeps required duration and true if err is FLOOD_WAIT
// or false and context or original error otherwise.
func FloodWait(ctx context.Context, err error) (bool, error) {
	if d, ok := mtproto.AsFloodWait(err); ok {
		select {
		case <-time.After(d + 1*time.Second):
			return true, err
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	return false, err
}
