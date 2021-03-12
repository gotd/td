package lifetime

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/xerrors"
)

// errRunnerStopped says that the runner was stopped by calling lifetime.Stop() func.
var errRunnerStopped = xerrors.Errorf("runner was stopped")

// Life represents a runner life.
type Life struct {
	waiters []func(e error)
	stop    func()
	err     error
	stopped bool
	mux     sync.Mutex
}

func (l *Life) waiter() func() error {
	ch := make(chan error)
	l.waiters = append(l.waiters, func(e error) { ch <- e })
	return func() error { return <-ch }
}

// Wait waits until life ends.
func (l *Life) Wait() error {
	l.mux.Lock()
	if l.stopped {
		l.mux.Unlock()
		return l.err
	}

	wait := l.waiter()
	l.mux.Unlock()
	return wait()
}

// Stop stops the life.
func (l *Life) Stop() error {
	l.mux.Lock()
	if l.stopped {
		l.mux.Unlock()
		return l.err
	}

	l.stop()
	wait := l.waiter()
	l.mux.Unlock()
	return wait()
}

// Start starts the runner.
func Start(r Runner) (*Life, error) {
	var (
		runResult = make(chan error)
		started   = make(chan struct{})
		stopper   = make(chan struct{})
		once      sync.Once
	)

	go func() {
		err := r.Run(context.Background(), func(ctx context.Context) error {
			close(started)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stopper:
				return errRunnerStopped
			}
		})
		runResult <- err
		close(runResult)
	}()

	select {
	case <-started:
		l := &Life{
			stop: func() { once.Do(func() { close(stopper) }) },
		}

		go func() {
			err := <-runResult
			if errors.Is(err, errRunnerStopped) {
				err = nil
			}

			l.mux.Lock()
			defer l.mux.Unlock()
			l.err = err
			l.stopped = true
			for _, cb := range l.waiters {
				cb(err)
			}
			l.waiters = nil
		}()

		return l, nil
	case err := <-runResult:
		return nil, err
	}
}
