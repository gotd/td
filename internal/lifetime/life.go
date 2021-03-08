package lifetime

import (
	"context"
	"errors"

	"golang.org/x/xerrors"
)

// ErrRunnerStopped says that the runner was stopped by calling lifetime.Stop() func.
var ErrRunnerStopped = xerrors.Errorf("runner was stopped")

// Life represents a runner life.
type Life struct {
	result chan error
	stop   func()
}

// Start starts the runner.
func Start(r Runner) (Life, error) {
	var (
		runResult = make(chan error)
		started   = make(chan struct{})
		stopper   = make(chan struct{})
	)

	go func() {
		err := r.Run(context.Background(), func(ctx context.Context) error {
			close(started)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stopper:
				return ErrRunnerStopped
			}
		})
		runResult <- err
		close(runResult)
	}()

	select {
	case <-started:
		return Life{
			result: runResult,
			stop:   func() { close(stopper) },
		}, nil
	case err := <-runResult:
		return Life{}, err
	}
}

// Stop stops the runner.
func Stop(l Life) error {
	defer close(l.result)

	l.stop()
	err := <-l.result
	if errors.Is(err, ErrRunnerStopped) {
		return nil
	}

	return err
}
