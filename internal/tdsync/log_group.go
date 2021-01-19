package tdsync

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/clock"
)

// LogGroup is simple wrapper around
// errgroup.Group to log task state.
// Unlike WaitGroup and errgroup.Group this is not allowed to use zero value.
type LogGroup struct {
	grp  *errgroup.Group
	gCtx context.Context

	log   *zap.Logger
	clock clock.Clock
}

// NewLogGroup creates new LogGroup.
func NewLogGroup(ctx context.Context, log *zap.Logger) *LogGroup {
	grp, gCtx := errgroup.WithContext(ctx)
	return &LogGroup{
		grp:   grp,
		gCtx:  gCtx,
		log:   log,
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
func (g *LogGroup) Go(taskName string, f func(ctx context.Context) error) {
	g.grp.Go(func() error {
		start := g.clock.Now()
		l := g.log.With(zap.String("task", taskName)).WithOptions(zap.AddCallerSkip(1))
		l.Debug("Task started")

		if err := f(g.gCtx); err != nil {
			elapsed := g.clock.Now().Sub(start)
			l.Debug("Task stopped", zap.Error(err), zap.Duration("elapsed", elapsed))
			return xerrors.Errorf("task %s: %w", taskName, err)
		}

		elapsed := g.clock.Now().Sub(start)
		l.Debug("Task complete", zap.Duration("elapsed", elapsed))
		return nil
	})
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *LogGroup) Wait() error {
	return g.grp.Wait()
}
