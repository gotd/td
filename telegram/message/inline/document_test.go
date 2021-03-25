package inline

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestDocument(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)
	doc := &tg.InputDocument{ID: 10, AccessHash: 10, FileReference: []byte{10}}

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal(int64(10), v.QueryID)

		for i := range v.Results {
			r, ok := v.Results[i].(*tg.InputBotInlineResultDocument)
			mock.True(ok)
			mock.NotEmpty(r.ID)
			mock.Equal(doc, r.Document)
			mock.Equal(r.Title, r.Type)
			mock.Equal(r.Description, r.Title)
		}
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Video(doc, MessageText("video")).Title(VideoType).
			Description(VideoType),
		File(doc, MessageText("file")).ID("10").Title(DocumentType).
			Description(DocumentType),
		Audio(doc, MessageText("audio")).ID("10").Title(AudioType).
			Description(AudioType),
		GIF(doc, MessageText("gif")).ID("10").Title(GIFType).
			Description(GIFType),
		MPEG4GIF(doc, MessageText("mpeg4gif")).ID("10").Title(MPEG4GIFType).
			Description(MPEG4GIFType),
		Voice(doc, MessageText("voice")).ID("10").Title(VoiceType).
			Description(VoiceType),
		Sticker(doc, MessageText("sticker")).ID("10").Title(StickerType).
			Description(StickerType),
	)
	mock.NoError(err)

	mock.Expect().ThenRPCErr(testRPCError())
	_, err = builder.Set(ctx,
		Video(doc, MessageText("video")).Title(VideoType).
			Description(VideoType),
	)
	mock.Error(err)
}
