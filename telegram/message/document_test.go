package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestDocument(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputDocument{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaDocument{ID: loc}, mock)
	expectSendMedia(t, &tg.InputMediaDocument{
		ID:         loc,
		TTLSeconds: 10,
		Query:      "10",
	}, mock)

	_, err := sender.Self().Document(ctx, loc)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, Document(loc).
		TTL(10*time.Second).Query("10"))
	require.NoError(t, err)
}

func TestDocumentExternal(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectSendMedia(t, &tg.InputMediaDocumentExternal{URL: "https://google.com"}, mock)
	expectSendMedia(t, &tg.InputMediaDocumentExternal{
		URL:        "https://github.com",
		TTLSeconds: 10,
	}, mock)

	_, err := sender.Self().DocumentExternal(ctx, "https://google.com")
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, DocumentExternal("https://github.com").TTL(10*time.Second))
	require.NoError(t, err)
}

func TestDocumentByHash(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	doc := &tg.Document{
		ID:            10,
		FileReference: []byte{10},
	}
	loc := new(tg.InputDocument)
	loc.FillFrom(doc)

	hash := []byte{1, 2, 3}
	size := 10
	mime := "rustmustdie"

	mock.ExpectCall(&tg.MessagesGetDocumentByHashRequest{
		SHA256:   hash,
		Size:     size,
		MimeType: mime,
	}).ThenResult(doc)
	expectSendMedia(t, &tg.InputMediaDocument{ID: loc}, mock)
	_, err := sender.Self().DocumentByHash(ctx, hash, size, mime)
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesGetDocumentByHashRequest{
		SHA256:   hash,
		Size:     size,
		MimeType: mime,
	}).ThenResult(doc)
	expectSendMedia(t, &tg.InputMediaDocument{
		ID:         loc,
		TTLSeconds: 10,
		Query:      "10",
	}, mock)
	_, err = sender.Self().Media(ctx, DocumentByHash(hash, size, mime).
		TTL(10*time.Second).Query("10"))
	require.NoError(t, err)
}

func TestUploadedDocument(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}
	loc := &tg.InputDocumentFileLocation{
		ID: 10,
	}

	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:      file,
		ForceFile: true,
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File: file,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{FileName: "abc.jpg"},
		},
		TTLSeconds: 10,
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:  file,
		Thumb: file,
		Stickers: []tg.InputDocumentClass{&tg.InputDocument{
			ID: loc.GetID(),
		}},
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeHasStickers{},
		},
	}, mock)

	_, err := sender.Self().File(ctx, file)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, UploadedDocument(file).TTL(10*time.Second).
		Filename("abc.jpg"))
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, UploadedDocument(file).Thumb(file).Stickers(loc).HasStickers())
	require.NoError(t, err)
}
