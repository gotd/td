package tdsync

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/clock"
)

// LogGroup is simple wrapper around CancellableGroup to log task state.
// Unlike WaitGroup and errgroup.Group this is not allowed to use zero value.
type LogGroup struct {
	group CancellableGroup

	log   log.Logger
	clock clock.Clock
}

// NewLogGroup creates new LogGroup.
func NewLogGroup(parent context.Context, logger log.Logger) *LogGroup {
	return &LogGroup{
		group: *NewCancellableGroup(parent),
		log:   log.OrNop(logger),
		clock: clock.System,
	}
}

// SetClock sets Clock to use.
func (g *LogGroup) SetClock(c clock.Clock) {
	g.clock = c
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *LogGroup) Go(taskName string, f func(groupCtx context.Context) error) {
	g.group.Go(func(ctx context.Context) error {
		start := g.clock.Now()
		l := log.For(g.log).With(log.String("task", taskName))
		l.Debug(ctx, "Task started")

		if err := f(ctx); err != nil {
			elapsed := g.clock.Now().Sub(start)
			l.Debug(ctx, "Task stopped", log.Error(err), log.Duration("elapsed", elapsed))
			return errors.Wrapf(err, "task %s", taskName)
		}

		elapsed := g.clock.Now().Sub(start)
		l.Debug(ctx, "Task complete", log.Duration("elapsed", elapsed))
		return nil
	})
}

// Cancel cancels all goroutines in group.
//
// Note: context cancellation error will be returned by Wait().
func (g *LogGroup) Cancel() {
	g.group.Cancel()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *LogGroup) Wait() error {
	return g.group.Wait()
}
