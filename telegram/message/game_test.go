package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestGame(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	game := &tg.InputGameID{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaGame{
		ID: game,
	}, mock)
	_, err := sender.Self().Media(ctx, Game(game))
	require.NoError(t, err)
}
