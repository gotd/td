package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWaitReady(t *testing.T) {
	t.Run("ReadyAndCanceled", func(t *testing.T) {
		// Both channels are ready: readiness must always win.
		ready := make(chan struct{})
		close(ready)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		for i := 0; i < 1000; i++ {
			require.True(t, waitReady(ctx, ready))
		}
	})
	t.Run("Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		require.False(t, waitReady(ctx, make(chan struct{})))
	})
	t.Run("Ready", func(t *testing.T) {
		ready := make(chan struct{})
		close(ready)
		require.True(t, waitReady(context.Background(), ready))
	})
}
