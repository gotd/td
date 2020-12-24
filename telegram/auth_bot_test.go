package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestClient_AuthBot(t *testing.T) {
	const token = "12345:token"
	t.Run("AuthAuthorization", func(t *testing.T) {
		require.NoError(t, newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
			assert.Equal(t, &tg.AuthImportBotAuthorizationRequest{
				BotAuthToken: token,
				APIID:        TestAppID,
				APIHash:      TestAppHash,
			}, body)
			u := &tg.User{}
			u.SetBot(true)
			return &tg.AuthAuthorization{User: u}, nil
		}).AuthBot(context.Background(), token))
	})
	t.Run("AuthAuthorizationSignUpRequired", func(t *testing.T) {
		require.Error(t, newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
			return &tg.AuthAuthorizationSignUpRequired{}, nil
		}).AuthBot(context.Background(), token))
	})
}
