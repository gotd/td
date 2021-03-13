package lifetime

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

// Runner interface.
type Runner interface {
	Run(ctx context.Context, f func(context.Context) error) error
}

// Manager helps to manage multiple runners.
// It's like an errgroup for runners with Stop() functionality.
type Manager struct {
	runners map[Runner]*Life
	g       errgroup.Group
	mux     sync.Mutex
}

// NewManager creates new lifetime manager.
func NewManager() *Manager {
	return &Manager{
		runners: map[Runner]*Life{},
	}
}

// Start tries to start runner.
// If the start fails, the other runners will be stopped
// and the error will be returned (in Wait() func too).
// You can stop the runner by using Stop() func.
func (m *Manager) Start(r Runner) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.runners[r]; ok {
		return nil
	}

	life, err := Start(r)
	if err != nil {
		err = xerrors.Errorf("start runner: %w", err)
		// Shutdown all other runners.
		m.g.Go(func() error { return err })
		return err
	}

	m.g.Go(life.Wait)
	m.runners[r] = life
	return nil
}

// Stop stops the runner.
// If runner stops normally - it does not impact to other runners.
// Otherwise, all other runners will be stopped and Wait() func returned an error.
func (m *Manager) Stop(r Runner) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	life, ok := m.runners[r]
	if !ok {
		return xerrors.Errorf("runner not found")
	}

	delete(m.runners, r)
	life.Stop()
	return life.Wait()
}

// Go is equivalent to errgroup's Go() func.
func (m *Manager) Go(f func() error) {
	m.g.Go(f)
}

// Wait waits for one of the runners to return an error
// (at startup, at work or shutdown stage) and returns this error.
// All other runners will be stopped.
func (m *Manager) Wait() error {
	defer m.Close()

	return m.g.Wait()
}

func (m *Manager) Close() {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, life := range m.runners {
		life.Stop()
	}
	m.runners = map[Runner]*Life{}
}
