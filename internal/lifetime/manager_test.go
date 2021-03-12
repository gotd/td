package lifetime_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/internal/lifetime"
	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	onStart, onExit func()
	shouldFail      bool
}

func (m *mockRunner) Run(ctx context.Context, f func(context.Context) error) error {
	if m.shouldFail {
		return fmt.Errorf("fail")
	}
	if m.onStart == nil {
		m.onStart = func() {}
	}
	if m.onExit == nil {
		m.onExit = func() {}
	}

	m.onStart()
	defer m.onExit()
	return f(ctx)
}

func TestManager(t *testing.T) {
	m := lifetime.NewManager()

	started, stopped := false, false
	r := &mockRunner{
		onStart: func() { started = true },
		onExit:  func() { stopped = true },
	}

	require.NoError(t, m.Start(r))
	require.Equal(t, true, started)

	require.NoError(t, m.Stop(r))
	require.Equal(t, true, stopped)

	wgerr := make(chan error)
	go func() { wgerr <- m.Wait(context.TODO()) }()

	require.Eventually(t, func() bool {
		select {
		case err := <-wgerr:
			require.NoError(t, err)
			return true
		default:
			return false
		}
	}, time.Millisecond*10, time.Millisecond)
}
