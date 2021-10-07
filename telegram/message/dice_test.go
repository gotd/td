package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestMediaDice(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectDice := func(emoticon string) {
		expectSendMedia(t, &tg.InputMediaDice{Emoticon: emoticon}, mock)
	}

	expectDice(DiceEmoticon)
	expectDice(DartsEmoticon)
	expectDice(BasketballEmoticon)
	expectDice(FootballEmoticon)
	expectDice(CasinoEmoticon)
	expectDice(BowlingEmoticon)

	_, err := sender.Self().Dice(ctx)
	require.NoError(t, err)
	_, err = sender.Self().Darts(ctx)
	require.NoError(t, err)
	_, err = sender.Self().Basketball(ctx)
	require.NoError(t, err)
	_, err = sender.Self().Football(ctx)
	require.NoError(t, err)
	_, err = sender.Self().Casino(ctx)
	require.NoError(t, err)
	_, err = sender.Self().Bowling(ctx)
	require.NoError(t, err)
}
