package lifetime

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

var ErrClosed = xerrors.Errorf("Lifetimer was closed")

var errStopRunner = xerrors.Errorf("kek")

type Runner interface {
	Run(ctx context.Context, f func(context.Context) error) error
}

type Lifetimer struct {
	runners map[Runner]life
	g       errgroup.Group
	closed  bool
	mux     sync.Mutex
}

type life struct {
	result chan error
	stop   func()
}

func New() *Lifetimer {
	return &Lifetimer{
		runners: map[Runner]life{},
	}
}

func (lf *Lifetimer) Start(r Runner) error {
	lf.mux.Lock()
	defer lf.mux.Unlock()

	if lf.closed {
		return ErrClosed
	}

	if _, ok := lf.runners[r]; ok {
		return nil
	}

	var (
		started   = make(chan struct{})
		runResult = make(chan error)
		stopper   = make(chan struct{})
	)

	// TODO(ccln): make sure that runner was successfully started
	// before putting them to 'g'
	lf.g.Go(func() error {
		err := r.Run(context.Background(), func(ctx context.Context) error {
			close(started)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stopper:
				return errStopRunner
			}
		})

		if errors.Is(err, errStopRunner) {
			err = nil
		}

		runResult <- err
		return err
	})

	select {
	case <-started:
		lf.runners[r] = life{
			result: runResult,
			stop:   func() { close(stopper) },
		}
		return nil
	case err := <-runResult:
		return err
	}
}

func (lf *Lifetimer) Stop(r Runner) error {
	lf.mux.Lock()
	defer lf.mux.Unlock()

	if lf.closed {
		return ErrClosed
	}

	life, ok := lf.runners[r]
	if !ok {
		return xerrors.Errorf("not found")
	}

	life.stop()
	return <-life.result
}

func (lf *Lifetimer) Wait(ctx context.Context) error {
	lf.mux.Lock()
	if lf.closed {
		lf.mux.Unlock()
		return ErrClosed
	}
	lf.mux.Unlock()

	defer func() {
		lf.mux.Lock()
		defer lf.mux.Unlock()

		lf.closed = true

		for _, life := range lf.runners {
			life.stop()
			_ = <-life.result
		}
	}()

	return lf.g.Wait()
}
