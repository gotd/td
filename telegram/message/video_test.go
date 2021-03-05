package message

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestRoundVideo(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultVideoMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeVideo{
				RoundMessage: true,
			},
		},
	}, mock)
	expectSendMedia(&tg.InputMediaUploadedDocument{
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

	mock.NoError(sender.Self().RoundVideo(ctx, file))
	mock.NoError(sender.Self().Media(ctx, RoundVideo(file).
		Duration(10*time.Second).
		Resolution(10, 10).
		SupportsStreaming(),
	))
}
