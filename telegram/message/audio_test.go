package message

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestVoice(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultVoiceMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAudio{
				Voice: true,
			},
		},
	}, mock)
	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultAudioMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAudio{
				Duration:  10,
				Title:     "Big Iron",
				Performer: "Marty Robbins",
				Waveform:  []byte{10},
			},
		},
	}, mock)

	mock.NoError(sender.Self().Voice(ctx, file))
	mock.NoError(sender.Self().Media(ctx, Audio(file).
		Duration(10*time.Second).
		Title("Big Iron").
		Performer("Marty Robbins").
		Waveform([]byte{10}),
	))
}
