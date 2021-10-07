package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func TestClient_AuthStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("Authorized", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		user := &tg.User{
			Username: "user",
		}
		mock.Expect().ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{user}})

		status, err := testClient(mock).Status(ctx)
		require.NoError(t, err)
		require.True(t, status.Authorized)
		require.Equal(t, user, status.User)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		mock.Expect().ThenUnregistered()

		status, err := testClient(mock).Status(ctx)
		require.NoError(t, err)
		require.False(t, status.Authorized)
	})

	t.Run("Error", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		mock.Expect().ThenRPCErr(&tgerr.Error{
			Code:    500,
			Message: "BRUH",
			Type:    "BRUH",
		})

		_, err := testClient(mock).Status(ctx)
		require.Error(t, err)
	})
}

func TestClient_AuthIfNecessary(t *testing.T) {
	ctx := context.Background()

	t.Run("Authorized", func(t *testing.T) {
		mock := tgmock.NewRequire(t)
		testUser := &tg.User{
			Username: "user",
		}
		mock.Expect().ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{testUser}})

		// Pass empty AuthFlow because it should not be called anyway.
		require.NoError(t, testClient(mock).IfNecessary(ctx, Flow{}))
	})
}
