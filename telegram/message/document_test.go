package message

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestDocument(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputDocument{
		ID: 10,
	}

	expectSendMedia(&tg.InputMediaDocument{ID: loc}, mock)
	expectSendMedia(&tg.InputMediaDocument{
		ID:         loc,
		TTLSeconds: 10,
		Query:      "10",
	}, mock)

	mock.NoError(sender.Self().Document(ctx, loc))
	mock.NoError(sender.Self().Media(ctx, Document(loc).
		TTL(10*time.Second).Query("10")))
}

func TestDocumentExternal(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectSendMedia(&tg.InputMediaDocumentExternal{URL: "https://google.com"}, mock)
	expectSendMedia(&tg.InputMediaDocumentExternal{
		URL:        "https://github.com",
		TTLSeconds: 10,
	}, mock)

	mock.NoError(sender.Self().DocumentExternal(ctx, "https://google.com"))
	mock.NoError(sender.Self().Media(ctx, DocumentExternal("https://github.com").TTL(10*time.Second)))
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

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      file,
		ForceFile: true,
	}, mock)
	expectSendMedia(&tg.InputMediaUploadedDocument{
		File: file,
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{FileName: "abc.jpg"},
		},
		TTLSeconds: 10,
	}, mock)
	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:  file,
		Thumb: file,
		Stickers: []tg.InputDocumentClass{&tg.InputDocument{
			ID: loc.GetID(),
		}},
		Attributes: []tg.DocumentAttributeClass{
			&tg.DocumentAttributeHasStickers{},
		},
	}, mock)

	mock.NoError(sender.Self().File(ctx, file))
	mock.NoError(sender.Self().Media(ctx, UploadedDocument(file).TTL(10*time.Second).
		Filename("abc.jpg")))
	mock.NoError(sender.Self().Media(ctx, UploadedDocument(file).Thumb(file).Stickers(loc).HasStickers()))
}
