package inline

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/widget"
	"github.com/gotd/td/tg"
)

func TestMessageWidget(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)
	const html = `<div>widget</div>`
	button := markup.Callback("ok", []byte("ok"))

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		var decoded tg.MessagesSetInlineBotResultsRequest
		var buf bin.Buffer
		require.NoError(t, v.Encode(&buf))
		require.NoError(t, decoded.Decode(&buf))
		require.Len(t, decoded.Results, 2)

		r, ok := decoded.Results[0].(*tg.InputBotInlineResult)
		require.True(t, ok)
		require.Equal(t, "w", r.ID)
		m, ok := r.SendMessage.(*tg.InputBotInlineMessageWidget)
		require.True(t, ok)
		require.Equal(t, html, m.HTML)
		require.Equal(t, 100, m.Width)
		require.Equal(t, 50, m.Height)
		require.Equal(t, "pub", m.PublicData)
		require.Equal(t, "chat", m.ChatData)
		require.True(t, m.AllowQueries)
		require.True(t, m.SharedWidget)
		require.True(t, m.InvertMedia)
		require.Equal(t, "caption", m.Message)
		require.Len(t, m.Entities, 1)
		require.NotNil(t, m.ReplyMarkup)

		r, ok = decoded.Results[1].(*tg.InputBotInlineResult)
		require.True(t, ok)
		s, ok := r.SendMessage.(*tg.InputBotInlineMessageWidgetID)
		require.True(t, ok)
		require.Equal(t, int64(42), s.WidgetID)
		require.Equal(t, "c", s.ChatData)
		require.True(t, s.InvertMedia)
		require.Equal(t, "shared", s.Message)
		require.NotNil(t, s.ReplyMarkup)
	}).ThenTrue()
	_, err := builder.Set(ctx,
		Article("Widget", MessageWidget(html, styling.Bold("caption")).
			Size(100, 50).
			PublicData("pub").
			ChatData("chat").
			AllowQueries(true).
			Shared(true).
			InvertMedia(true).
			Row(button),
		).ID("w"),
		Article("Shared", MessageSharedWidget(42, styling.Plain("shared")).
			ChatData("c").
			InvertMedia(true).
			Row(button),
		),
	)
	require.NoError(t, err)

	// The compressed size limit is enforced by the server only.
	var sb strings.Builder
	r := rand.New(rand.NewPCG(1, 2)) // #nosec G404
	for range 8000 {
		sb.WriteByte(byte('a' + r.IntN(26)))
	}
	big := sb.String()
	require.ErrorIs(t, widget.ValidateCompressed(big), widget.ErrHTMLCompressedTooLong)
	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		r := v.Results[0].(*tg.InputBotInlineResult)
		require.Equal(t, big, r.SendMessage.(*tg.InputBotInlineMessageWidget).HTML)
	}).ThenTrue()
	_, err = builder.Set(ctx, Article("Big", MessageWidget(big)))
	require.NoError(t, err)

	_, err = builder.Set(ctx, Article("Empty", MessageWidget("")))
	require.ErrorIs(t, err, widget.ErrEmptyHTML)
	_, err = builder.Set(ctx, Article("Size", MessageWidget(html).Size(0, -1)))
	require.ErrorIs(t, err, widget.ErrInvalidSize)
	_, err = builder.Set(ctx, Article("Shared", MessageSharedWidget(0)))
	require.ErrorIs(t, err, widget.ErrInvalidWidgetID)
}
