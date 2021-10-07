package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestGIF(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultGifMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAnimated{},
		},
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultGifMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAnimated{},
		},
		TTLSeconds: 10,
	}, mock)

	_, err := sender.Self().GIF(ctx, file)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, GIF(file).TTL(10*time.Second))
	require.NoError(t, err)
}
