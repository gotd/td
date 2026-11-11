package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// A caption without formatting must not send an empty entities vector: its
// flag would be set (flags.3) while Telegram Desktop omits the field.
func TestSendMediaPlainCaptionNoEntities(t *testing.T) {
	sender, mock := testSender(t)
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		require.Equal(t, "caption", req.Message)
		require.Nil(t, req.Entities)
		var buf bin.Buffer
		require.NoError(t, req.Encode(&buf))
		require.False(t, req.Flags.Has(3), "entities flag must not be set")
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Media(context.Background(),
		UploadedPhoto(&tg.InputFile{ID: 1, Parts: 1, Name: "a.jpg"}, styling.Plain("caption")))
	require.NoError(t, err)
}

func TestEditPlainTextNoEntities(t *testing.T) {
	sender, mock := testSender(t)
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesEditMessageRequest)
		require.True(t, ok)
		require.Nil(t, req.Entities)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Edit(10).StyledText(context.Background(), styling.Plain("text"))
	require.NoError(t, err)
}
