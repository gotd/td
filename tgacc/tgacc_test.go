package tgacc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/testutil"
)

func TestExternalE2E(t *testing.T) {
	testutil.SkipExternal(t)

	manager, err := NewTestAccountManager()
	require.NoError(t, err)

	ctx := context.Background()
	client, err := manager.Acquire(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
}
