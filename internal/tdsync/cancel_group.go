// Package tdsync contains some useful synchronization utilities.
package tdsync

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// CancellableGroup is simple wrapper around
// errgroup.Group to make group cancellation easier.
// Unlike WaitGroup and errgroup.Group this is not allowed to use zero value.
type CancellableGroup struct {
	cancel context.CancelFunc

	group *errgroup.Group
	ctx   context.Context
}

// NewCancellableGroup creates new CancellableGroup.
//
// Example:
//
//	g := NewCancellableGroup(ctx)
//	g.Go(func(ctx context.Context) error {
//		<-ctx.Done()
//		return ctx.Err()
//	})
//	g.Cancel()
//	g.Wait()
func NewCancellableGroup(parent context.Context) *CancellableGroup {
	ctx, cancel := context.WithCancel(parent)
	group, groupCtx := errgroup.WithContext(ctx)

	return &CancellableGroup{
		cancel: cancel,
		group:  group,
		ctx:    groupCtx,
	}
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *CancellableGroup) Go(f func(ctx context.Context) error) {
	g.group.Go(func() error {
		return f(g.ctx)
	})
}

// Cancel cancels all goroutines in group.
//
// Note: context cancellation error will be returned by Wait().
func (g *CancellableGroup) Cancel() {
	g.cancel()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *CancellableGroup) Wait() error {
	return g.group.Wait()
}
