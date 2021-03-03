package message

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestMediaDice(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectDice := func(emoticon string) {
		expectSendMedia(&tg.InputMediaDice{Emoticon: emoticon}, mock)
	}

	expectDice(DiceEmoticon)
	expectDice(DartsEmoticon)
	expectDice(BasketballEmoticon)

	mock.NoError(sender.Self().Dice(ctx))
	mock.NoError(sender.Self().Darts(ctx))
	mock.NoError(sender.Self().Basketball(ctx))
}
