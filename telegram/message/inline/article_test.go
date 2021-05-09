package inline

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestArticle(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal(int64(10), v.QueryID)

		for i := range v.Results {
			r, ok := v.Results[i].(*tg.InputBotInlineResult)
			mock.True(ok)
			mock.NotEmpty(r.ID)
			mock.Equal(r.Title, r.Type)
			mock.Equal(r.Description, r.Title)
			mock.Equal(r.URL, r.Description)
		}
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Article(ArticleType, MessageText("article")).
			Description(ArticleType).URL(ArticleType),
		Article(ArticleType, MediaAuto("article")).ID("10").Title(ArticleType).
			Description(ArticleType).URL(ArticleType),
	)
	mock.NoError(err)

	mock.Expect().ThenRPCErr(testRPCError())
	_, err = builder.Set(ctx,
		Article(ArticleType, MessageText("article")),
	)
	mock.Error(err)
}
