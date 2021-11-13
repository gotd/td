package tgerr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gotd/neo"
)

func TestFloodWait(t *testing.T) {
	floodWaitErr := New(400, ErrFloodWait)
	floodWaitErr.Argument = 3600
	ctx := context.Background()

	t.Run("WrongType", func(t *testing.T) {
		a := assert.New(t)

		e := New(0, "wrong")
		ok, err := FloodWait(ctx, e)
		a.False(ok)
		a.ErrorIs(err, e)
	})
	t.Run("Wait", func(t *testing.T) {
		a := assert.New(t)
		c := neo.NewTime(time.Now())
		e := floodWaitErr

		done := make(chan struct{})
		observer := c.Observe()
		go func() {
			ok, err := FloodWait(ctx, e, FloodWaitWithClock(c))
			a.True(ok)
			a.ErrorIs(err, e)

			close(done)
		}()

		<-observer
		c.Travel(3601 * time.Second)
		<-done
	})
	t.Run("ContextDone", func(t *testing.T) {
		a := assert.New(t)
		c := neo.NewTime(time.Now())
		e := floodWaitErr

		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		ok, err := FloodWait(canceledCtx, e, FloodWaitWithClock(c))
		a.False(ok)
		a.ErrorIs(err, context.Canceled)
	})
}
