package message

import (
	"context"
	"io"
	"io/fs"
	"testing"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/telegram/uploader/source"
	"github.com/gotd/td/tg"
)

type mockUploader struct {
	file tg.InputFileClass
}

func (m mockUploader) FromFile(ctx context.Context, f uploader.File) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromPath(ctx context.Context, path string) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromFS(ctx context.Context, filesystem fs.FS, path string) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromReader(ctx context.Context, name string, f io.Reader) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromBytes(ctx context.Context, name string, b []byte) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromURL(ctx context.Context, rawURL string) (tg.InputFileClass, error) {
	return m.file, nil
}

func (m mockUploader) FromSource(ctx context.Context, src source.Source, rawURL string) (tg.InputFileClass, error) {
	return m.file, nil
}

func TestUpload(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	f := &tg.InputFile{
		ID:          1,
		Parts:       1,
		Name:        "10.jpg",
		MD5Checksum: "abc",
	}
	upd := mockUploader{file: f}
	dialog := sender.WithUploader(upd).Self()

	expectSendMedia(&tg.InputMediaUploadedPhoto{
		File: f,
	}, mock)
	_, err := dialog.Upload(FromPath("abc.jpg")).Photo(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromReader("abc.jpg", nil)).File(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromFS(nil, "abc.jpg")).File(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromBytes("abc.jpg", nil)).File(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromFile(nil)).File(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromURL("http://example.com")).File(ctx)
	mock.NoError(err)

	expectSendMedia(&tg.InputMediaUploadedDocument{
		File:      f,
		ForceFile: true,
	}, mock)
	_, err = dialog.Upload(FromSource(source.NewHTTPSource(), "http://example.com")).File(ctx)
	mock.NoError(err)
}
