package inline

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestGame(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)
	gameName := "game"

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal(int64(10), v.QueryID)

		for i := range v.Results {
			r, ok := v.Results[i].(*tg.InputBotInlineResultGame)
			mock.True(ok)
			mock.NotEmpty(r.ID)
			mock.Equal(gameName, r.ShortName)
		}
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Game(gameName, MessageText("game")),
		Game(gameName, MessageText("game")).ID("10"),
	)
	mock.NoError(err)

	mock.Expect().ThenRPCErr(testRPCError())
	_, err = builder.Set(ctx,
		Game(gameName, MessageText("game")),
		Game(gameName, MessageText("game")).ID("10"),
	)
	mock.Error(err)
}
