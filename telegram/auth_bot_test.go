package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestClient_AuthBot(t *testing.T) {
	const token = "12345:token"

	t.Run("AuthAuthorization", func(t *testing.T) {
		inv := mockInvoker(func(input *bin.Buffer) (bin.Encoder, error) {
			var req tg.AuthImportBotAuthorizationRequest
			if err := req.Decode(input); err != nil {
				return nil, err
			}

			require.Equal(t, &tg.AuthImportBotAuthorizationRequest{
				BotAuthToken: token,
				APIID:        1,
				APIHash:      "hash",
			}, &req)

			u := &tg.User{}
			u.SetBot(true)
			return &tg.AuthAuthorization{User: u}, nil
		})

		client := &Client{
			RPC:     tg.NewClient(inv),
			mtp:     inv,
			appID:   1,
			appHash: "hash",
		}

		require.NoError(t, client.AuthBot(context.Background(), token))
	})
	t.Run("AuthAuthorizationSignUpRequired", func(t *testing.T) {
		inv := mockInvoker(func(input *bin.Buffer) (bin.Encoder, error) {
			var req tg.AuthImportBotAuthorizationRequest
			if err := req.Decode(input); err != nil {
				return nil, err
			}

			return &tg.AuthAuthorizationSignUpRequired{}, nil
		})

		client := &Client{
			RPC:     tg.NewClient(inv),
			appID:   1,
			appHash: "hash",
		}

		require.Error(t, client.AuthBot(context.Background(), token))
	})
}
