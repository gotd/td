package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// PEER-5: Manager serves user access hashes to the updates manager, only for
// its own account like the channel hashes.
func TestManagerUserAccessHasher(t *testing.T) {
	ctx := context.Background()
	_, m := testManager(t)

	// Unknown self: ignored.
	require.NoError(t, m.SetUserAccessHash(ctx, 1, 10, 777))
	_, found, err := m.GetUserAccessHash(ctx, 1, 10)
	require.NoError(t, err)
	require.False(t, found)

	self := getTestSelf()
	m.me.Store(self)
	require.NoError(t, m.SetUserAccessHash(ctx, self.ID, 10, 777))
	hash, found, err := m.GetUserAccessHash(ctx, self.ID, 10)
	require.NoError(t, err)
	require.True(t, found)
	require.EqualValues(t, 777, hash)

	// Another account's hashes are not served.
	_, found, err = m.GetUserAccessHash(ctx, self.ID+1, 10)
	require.NoError(t, err)
	require.False(t, found)

	// Hashes of applied users are served too.
	u := getTestUser()
	require.NoError(t, m.Apply(ctx, nil, nil))
	require.NoError(t, m.applyUsers(ctx, u))
	hash, found, err = m.GetUserAccessHash(ctx, self.ID, u.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, u.AccessHash, hash)
}
