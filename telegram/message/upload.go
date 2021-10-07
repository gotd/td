package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// Upload creates new UploadBuilder to upload and send attachments.
// Given option will be called only once, even if you call upload functions.
func (b *Builder) Upload(upd UploadOption) *UploadBuilder {
	return &UploadBuilder{
		builder: b.copy(),
		option:  upd,
	}
}

// UploadBuilder is an attachment uploading helper.
type UploadBuilder struct {
	builder *Builder
	option  UploadOption
}

func (u *UploadBuilder) file(ctx context.Context) (tg.InputFileClass, error) {
	return u.option.apply(ctx, uploadBuilder{upload: u.builder.sender.uploader})
}

// AsInputFile uploads and returns uploaded file entity.
func (u *UploadBuilder) AsInputFile(ctx context.Context) (tg.InputFileClass, error) {
	return u.file(ctx)
}

func (u *UploadBuilder) send(
	ctx context.Context,
	using func(ctx context.Context, file tg.InputFileClass, caption ...StyledTextOption) (tg.UpdatesClass, error),
	caption []StyledTextOption,
) (tg.UpdatesClass, error) {
	f, err := u.file(ctx)
	if err != nil {
		return nil, xerrors.Errorf("upload: %w", err)
	}

	return using(ctx, f, caption...)
}

// Photo uploads and sends photo.
func (u *UploadBuilder) Photo(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.UploadedPhoto, caption)
}

// Audio uploads and sends audio file.
func (u *UploadBuilder) Audio(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.Audio, caption)
}

// Voice uploads and sends voice message.
func (u *UploadBuilder) Voice(ctx context.Context) (tg.UpdatesClass, error) {
	f, err := u.file(ctx)
	if err != nil {
		return nil, xerrors.Errorf("upload: %w", err)
	}

	return u.builder.Voice(ctx, f)
}

// Video uploads and sends video.
func (u *UploadBuilder) Video(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.Video, caption)
}

// RoundVideo uploads and sends round video.
func (u *UploadBuilder) RoundVideo(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.RoundVideo, caption)
}

// GIF uploads and sends gif file.
func (u *UploadBuilder) GIF(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.GIF, caption)
}

// Sticker uploads and sends sticker.
func (u *UploadBuilder) Sticker(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.UploadedSticker, caption)
}

// File uploads and sends plain file.
func (u *UploadBuilder) File(ctx context.Context, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return u.send(ctx, u.builder.File, caption)
}
