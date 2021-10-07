package inline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

func TestArticle(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		require.Equal(t, int64(10), v.QueryID)

		for i := range v.Results {
			r, ok := v.Results[i].(*tg.InputBotInlineResult)
			require.True(t, ok)
			require.NotZero(t, r.ID)
			require.Equal(t, r.Title, r.Type)
			require.Equal(t, r.Description, r.Title)
			require.Equal(t, r.URL, r.Description)
		}
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Article(ArticleType, MessageText("article")).
			Description(ArticleType).URL(ArticleType),
		Article(ArticleType, MediaAuto("article")).ID("10").Title(ArticleType).
			Description(ArticleType).URL(ArticleType),
	)
	require.NoError(t, err)

	mock.Expect().ThenRPCErr(testRPCError())
	_, err = builder.Set(ctx,
		Article(ArticleType, MessageText("article")),
	)
	require.Error(t, err)
}
