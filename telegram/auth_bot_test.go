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
		u := &tg.User{}
		u.SetBot(true)

		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorization{User: u})

		a.NoError(client.AuthBot(context.Background(), token))
	}))

	t.Run("AuthAuthorizationSignUpRequired", mockClient(func(a *rpcmock.Mock, client *Client) {
		a.ExpectCall(&tg.AuthImportBotAuthorizationRequest{
			BotAuthToken: token,
			APIID:        TestAppID,
			APIHash:      TestAppHash,
		}).ThenResult(&tg.AuthAuthorizationSignUpRequired{})

		a.Error(client.AuthBot(context.Background(), token))
	}))
}
