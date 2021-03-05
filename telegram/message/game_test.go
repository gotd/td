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
	mock.NoError(sender.Self().Media(ctx, Game(game)))
}
