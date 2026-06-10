package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestSpoiler(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	file := &tg.InputFile{ID: 10}
	photo := &tg.InputPhoto{ID: 10}
	doc := &tg.InputDocument{ID: 10}

	expectSendMedia(t, &tg.InputMediaUploadedPhoto{File: file, Spoiler: true}, mock)
	expectSendMedia(t, &tg.InputMediaPhoto{ID: photo, Spoiler: true}, mock)
	expectSendMedia(t, &tg.InputMediaPhotoExternal{URL: "https://google.com", Spoiler: true}, mock)
	expectSendMedia(t, &tg.InputMediaDocument{ID: doc, Spoiler: true}, mock)
	expectSendMedia(t, &tg.InputMediaDocumentExternal{URL: "https://google.com", Spoiler: true}, mock)

	_, err := sender.Self().Media(ctx, UploadedPhoto(file).Spoiler(true))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, Photo(photo).Spoiler(true))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, PhotoExternal("https://google.com").Spoiler(true))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, Document(doc).Spoiler(true))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, DocumentExternal("https://google.com").Spoiler(true))
	require.NoError(t, err)
}
