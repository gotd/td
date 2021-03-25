package inline

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestPhoto(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)
	photo := &tg.InputPhoto{
		ID:            10,
		AccessHash:    10,
		FileReference: []byte{10},
	}

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal(int64(10), v.QueryID)

		for i := range v.Results {
			r, ok := v.Results[i].(*tg.InputBotInlineResultPhoto)
			mock.True(ok)
			mock.NotEmpty(r.ID)
			mock.Equal(photo, r.Photo)
		}
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Photo(photo, MessageText("photo")),
		Photo(photo, MessageText("photo")).ID("10"),
	)
	mock.NoError(err)

	mock.Expect().ThenRPCErr(testRPCError())
	_, err = builder.Set(ctx,
		Photo(photo, MessageGeo(&tg.InputGeoPoint{
			Lat:            10,
			Long:           42,
			AccuracyRadius: 1337,
		})),
		Photo(photo, MessageGame()).ID("10"),
	)
	mock.Error(err)
}
