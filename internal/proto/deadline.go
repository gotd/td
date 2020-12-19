package proto

import (
	"context"
	"time"
)

const defaultTimeout = time.Second * 10

func deadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}
	return time.Now().Add(defaultTimeout)
}
