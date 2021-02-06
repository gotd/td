package tdsync

import (
	"context"
	"sync"
)

// Supervisor is simple task group primitive to control multiple
// long-live tasks.
// Unlike Groups, Supervisor does not cancel when one task is failed.
// Unlike WaitGroup and errgroup.Group this is not allowed to use zero value.
type Supervisor struct {
	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	onError func(err error)
}

// NewSupervisor creates new Supervisor.
func NewSupervisor(parent context.Context) *Supervisor {
	ctx, cancel := context.WithCancel(parent)

	return &Supervisor{
		ctx:    ctx,
		cancel: cancel,
	}
}

// WithErrorHandler sets tasks error handler
// Must be called before any Go calls.
func (s *Supervisor) WithErrorHandler(h func(err error)) *Supervisor {
	s.onError = h
	return s
}

// Go calls the given function in a new goroutine.
func (s *Supervisor) Go(task func(ctx context.Context) error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if err := task(s.ctx); err != nil {
			if s.onError != nil {
				s.onError(err)
			}
		}
	}()
}

// Cancel cancels all goroutines in group.
//
// Note: context cancellation error can be returned by Wait().
func (s *Supervisor) Cancel() {
	s.cancel()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (s *Supervisor) Wait() error {
	s.wg.Wait()
	return nil
}
