package message

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// PhotoBuilder is a Photo media option.
type PhotoBuilder struct {
	photo   tg.InputMediaPhoto
	caption []StyledTextOption
}

// TTL sets time to live of self-destructing photo.
func (u *PhotoBuilder) TTL(ttl time.Duration) *PhotoBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing photo.
func (u *PhotoBuilder) TTLSeconds(ttl int) *PhotoBuilder {
	u.photo.TTLSeconds = ttl
	return u
}

// apply implements MediaOption.
func (u *PhotoBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.photo, u.caption...).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *PhotoBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return u.apply(ctx, b)
}

// Photo adds photo attachment.
func Photo(photo FileLocation, caption ...StyledTextOption) *PhotoBuilder {
	v := new(tg.InputPhoto)
	v.FillFrom(photo)

	return &PhotoBuilder{
		photo: tg.InputMediaPhoto{
			ID: v,
		},
		caption: caption,
	}
}

// Photo sends photo.
func (b *Builder) Photo(
	ctx context.Context,
	photo FileLocation, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, Photo(photo, caption...))
}

// PhotoExternalBuilder is a PhotoExternal media option.
type PhotoExternalBuilder struct {
	doc     tg.InputMediaPhotoExternal
	caption []StyledTextOption
}

// TTL sets time to live of self-destructing document.
func (u *PhotoExternalBuilder) TTL(ttl time.Duration) *PhotoExternalBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing document.
func (u *PhotoExternalBuilder) TTLSeconds(ttl int) *PhotoExternalBuilder {
	u.doc.TTLSeconds = ttl
	return u
}

// apply implements MediaOption.
func (u *PhotoExternalBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.doc, u.caption...).apply(ctx, b)
}

// PhotoExternal adds photo attachment which will be downloaded by the Telegram servers.
func PhotoExternal(url string, caption ...StyledTextOption) *PhotoExternalBuilder {
	return &PhotoExternalBuilder{
		doc: tg.InputMediaPhotoExternal{
			URL: url,
		},
		caption: caption,
	}
}

// PhotoExternal sends photo attachment which will be downloaded by the Telegram servers.
func (b *Builder) PhotoExternal(
	ctx context.Context,
	url string, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, PhotoExternal(url, caption...))
}

// UploadedPhotoBuilder is a UploadedPhoto media option.
type UploadedPhotoBuilder struct {
	photo   tg.InputMediaUploadedPhoto
	caption []StyledTextOption
}

// Stickers adds attached mask stickers.
func (u *UploadedPhotoBuilder) Stickers(stickers ...FileLocation) *UploadedPhotoBuilder {
	u.photo.Stickers = append(u.photo.Stickers, inputDocuments(stickers...)...)
	return u
}

// TTL sets time to live of self-destructing photo.
func (u *UploadedPhotoBuilder) TTL(ttl time.Duration) *UploadedPhotoBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing photo.
func (u *UploadedPhotoBuilder) TTLSeconds(ttl int) *UploadedPhotoBuilder {
	u.photo.TTLSeconds = ttl
	return u
}

// apply implements MediaOption.
func (u *UploadedPhotoBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.photo, u.caption...).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *UploadedPhotoBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	m, err := b.sender.uploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  b.peer,
		Media: &u.photo,
	})
	if err != nil {
		return xerrors.Errorf("upload media: %w", err)
	}

	input, err := convertMessageMediaToInput(m)
	if err != nil {
		return xerrors.Errorf("convert: %w", err)
	}

	return Media(input, u.caption...).apply(ctx, b)
}

// UploadedPhoto adds photo attachment.
func UploadedPhoto(file tg.InputFileClass, caption ...StyledTextOption) *UploadedPhotoBuilder {
	return &UploadedPhotoBuilder{
		photo: tg.InputMediaUploadedPhoto{
			File: file,
		},
		caption: caption,
	}
}

// UploadedPhoto sends uploaded file as photo.
func (b *Builder) UploadedPhoto(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, UploadedPhoto(file, caption...))
}
