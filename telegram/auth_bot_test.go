package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestClient_AuthBot(t *testing.T) {
	const token = "12345:token"

	t.Run("AuthAuthorization", mockClient(func(a *tgmock.Mock, client *Client) {
		testUser := &tg.User{}
		testUser.SetBot(true)

		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorization{User: testUser})

		result, err := client.AuthBot(context.Background(), token)
		require.NoError(t, err)
		require.Equal(t, testUser, result.User)
	}))

	t.Run("AuthAuthorizationSignUpRequired", mockClient(func(a *tgmock.Mock, client *Client) {
		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorizationSignUpRequired{})

		result, err := client.AuthBot(context.Background(), token)
		require.Error(t, err)
		require.Nil(t, result)
	}))
}
