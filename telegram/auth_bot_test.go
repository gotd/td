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
	require.NoError(t, newTestClient(func(id int64, body bin.Encoder) (bin.Encoder, error) {
		assert.Equal(t, &tg.AuthImportBotAuthorizationRequest{BotAuthToken: token}, body)
		u := &tg.User{}
		u.SetBot(true)
		return &tg.AuthAuthorization{User: u}, nil
	}).AuthBot(context.Background(), token))
}
