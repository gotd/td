package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestVoice(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: DefaultVoiceMIME,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeAudio{
				Voice: true,
			},
		},
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
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

	_, err := sender.Self().Voice(ctx, file)
	require.NoError(t, err)

	_, err = sender.Self().Media(ctx, Audio(file).
		Duration(10*time.Second).
		Title("Big Iron").
		Performer("Marty Robbins").
		Waveform([]byte{10}),
	)
	require.NoError(t, err)
}
