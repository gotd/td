package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_Self(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)
	testUser := getTestSelf()

	_, ok := m.myID()
	a.False(ok)
	a.False(m.selfIsBot())

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{&tg.InputUserSelf{}},
	}).ThenRPCErr(getTestError())
	u, err := m.Self(ctx)
	a.Error(err)
	a.Zero(u)

	mock.ExpectCall(&tg.UsersGetUsersRequest{
		ID: []tg.InputUserClass{&tg.InputUserSelf{}},
	}).ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{testUser}})
	u, err = m.Self(ctx)
	a.NoError(err)
	a.Equal(testUser, u.Raw())

	// Test caching.
	u, err = m.Self(ctx)
	a.NoError(err)
	a.Equal(testUser, u.Raw())

	id, ok := m.myID()
	a.True(ok)
	a.Equal(testUser.ID, id)
	a.True(m.selfIsBot())
}
