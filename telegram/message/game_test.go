package message

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestGame(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	game := &tg.InputGameID{
		ID: 10,
	}

	expectSendMedia(&tg.InputMediaGame{
		ID: game,
	}, mock)
	_, err := sender.Self().Media(ctx, Game(game))
	mock.NoError(err)
}
