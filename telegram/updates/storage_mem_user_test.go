package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemUserAccessHasher(t *testing.T) {
	ctx := context.Background()
	h := newMemUserAccessHasher()

	_, found, err := h.GetUserAccessHash(ctx, 1, 555)
	require.NoError(t, err)
	require.False(t, found, "unset user must report not found")

	require.NoError(t, h.SetUserAccessHash(ctx, 1, 555, 7777))

	hash, found, err := h.GetUserAccessHash(ctx, 1, 555)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(7777), hash)

	// Scoped per self user id.
	_, found, err = h.GetUserAccessHash(ctx, 2, 555)
	require.NoError(t, err)
	require.False(t, found, "hash is scoped to the owning self user id")
}
