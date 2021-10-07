package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestRoundVideo(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultVideoMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeVideo{
				RoundMessage: true,
			},
		},
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultVideoMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeVideo{
				RoundMessage:      true,
				SupportsStreaming: true,
				Duration:          10,
				W:                 10,
				H:                 10,
			},
		},
	}, mock)

	_, err := sender.Self().RoundVideo(ctx, file)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, RoundVideo(file).
		Duration(10*time.Second).
		Resolution(10, 10).
		SupportsStreaming(),
	)
	require.NoError(t, err)
}
