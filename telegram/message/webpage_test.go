package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestBuilder_InvertMedia(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		require.Equal(t, "abc", req.Message)
		require.True(t, req.InvertMedia)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().InvertMedia().Text(ctx, "abc")
	require.NoError(t, err)
}

func TestWebPage(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectSendMedia(t, &tg.InputMediaWebPage{URL: "https://google.com"}, mock)
	expectSendMedia(t, &tg.InputMediaWebPage{
		URL:             "https://github.com",
		ForceLargeMedia: true,
		Optional:        true,
	}, mock)

	_, err := sender.Self().WebPage(ctx, "https://google.com")
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx,
		WebPage("https://github.com").ForceLargeMedia(true).Optional(true),
	)
	require.NoError(t, err)
}
