package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestPhoto(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputPhoto{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaPhoto{ID: loc}, mock)
	expectSendMedia(t, &tg.InputMediaPhoto{
		ID:         loc,
		TTLSeconds: 10,
	}, mock)

	_, err := sender.Self().Photo(ctx, loc)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, Photo(loc).TTL(10*time.Second))
	require.NoError(t, err)
}

func TestPhotoExternal(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectSendMedia(t, &tg.InputMediaPhotoExternal{URL: "https://google.com"}, mock)
	expectSendMedia(t, &tg.InputMediaPhotoExternal{
		URL:        "https://github.com",
		TTLSeconds: 10,
	}, mock)

	_, err := sender.Self().PhotoExternal(ctx, "https://google.com")
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, PhotoExternal("https://github.com").TTL(10*time.Second))
	require.NoError(t, err)
}

func TestUploadedPhoto(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}
	loc := &tg.InputDocumentFileLocation{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaUploadedPhoto{
		File: file,
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedPhoto{
		File:       file,
		TTLSeconds: 10,
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedPhoto{
		File: file,
		Stickers: []tg.InputDocumentClass{&tg.InputDocument{
			ID: loc.GetID(),
		}},
	}, mock)

	_, err := sender.Self().UploadedPhoto(ctx, file)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, UploadedPhoto(file).TTL(10*time.Second))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, UploadedPhoto(file).Stickers(loc))
	require.NoError(t, err)
}
