package telegram

import (
	"context"
	"testing"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func TestClient_AuthBot(t *testing.T) {
	const token = "12345:token"

	t.Run("AuthAuthorization", mockClient(func(a *rpcmock.Mock, client *Client) {
		testUser := &tg.User{}
		testUser.SetBot(true)

		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorization{User: testUser})

		result, err := client.AuthBot(context.Background(), token)
		a.NoError(err)
		a.Equal(testUser, result.User)
	}))

	t.Run("AuthAuthorizationSignUpRequired", mockClient(func(a *rpcmock.Mock, client *Client) {
		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorizationSignUpRequired{})

		result, err := client.AuthBot(context.Background(), token)
		a.Error(err)
		a.Nil(result)
	}))
}
