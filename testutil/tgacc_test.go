package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalE2E(t *testing.T) {
	SkipExternal(t)

	manager, err := NewTestAccountManager()
	require.NoError(t, err)

	ctx := context.Background()
	client, err := manager.Acquire(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
}
