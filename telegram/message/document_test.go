package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
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
	size := int64(10)
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
		MimeType:  "application/octet-stream",
	}, mock)
	expectSendMedia(t, &tg.InputMediaUploadedDocument{
		File:     file,
		MimeType: "image/jpeg",
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
		MimeType: "application/octet-stream",
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

// A file of unknown type is sent as application/octet-stream with its name,
// like Telegram Desktop; without both, clients showed it as an audio track.
func TestUploadedDocumentDefaults(t *testing.T) {
	for _, tt := range []struct {
		name, mime string
	}{
		{"cmp-doc.bin", "application/octet-stream"},
		{"noext", "application/octet-stream"},
		{"notes.unknownext", "application/octet-stream"},
		{"a.jpg", "image/jpeg"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := File(&tg.InputFile{ID: 1, Parts: 1, Name: tt.name}).media()
			require.Equal(t, tt.mime, doc.MimeType)
			require.Equal(t, []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: tt.name},
			}, doc.Attributes)
		})
	}

	// Explicit MIME and file name win.
	doc := File(&tg.InputFile{ID: 1, Name: "a.bin"}).MIME("text/csv").Filename("report.csv").media()
	require.Equal(t, "text/csv", doc.MimeType)
	require.Equal(t, []tg.DocumentAttributeClass{
		&tg.DocumentAttributeFilename{FileName: "report.csv"},
	}, doc.Attributes)
}
