package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func TestClient_AuthStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("Authorized", mockClient(func(a *tgmock.Mock, client *Client) {
		user := &tg.User{
			Username: "user",
		}
		a.Expect().ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{user}})

		status, err := client.AuthStatus(ctx)
		require.NoError(t, err)
		require.True(t, status.Authorized)
		require.Equal(t, user, status.User)
	}))

	t.Run("Unauthorized", mockClient(func(a *tgmock.Mock, client *Client) {
		a.Expect().ThenUnregistered()

		status, err := client.AuthStatus(ctx)
		require.NoError(t, err)
		require.False(t, status.Authorized)
	}))

	t.Run("Error", mockClient(func(a *tgmock.Mock, client *Client) {
		a.Expect().ThenRPCErr(&tgerr.Error{
			Code:    500,
			Message: "BRUH",
			Type:    "BRUH",
		})

		_, err := client.AuthStatus(ctx)
		require.Error(t, err)
	}))
}

func TestClient_AuthIfNecessary(t *testing.T) {
	ctx := context.Background()

	t.Run("Authorized", mockClient(func(a *tgmock.Mock, client *Client) {
		testUser := &tg.User{
			Username: "user",
		}
		a.Expect().ThenResult(&tg.UserClassVector{Elems: []tg.UserClass{testUser}})

		// Pass empty AuthFlow because it should not be called anyway.
		require.NoError(t, client.AuthIfNecessary(ctx, AuthFlow{}))
	}))
}
