package message

import (
	"context"
	"time"

	"github.com/gotd/td/tg"
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

// Apply implements MediaOption.
func (u *PhotoBuilder) Apply(ctx context.Context, b multiMediaBuilder) error {
	return Media(&u.photo, u.caption...).Apply(ctx, b)
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

// Apply implements MediaOption.
func (u *PhotoExternalBuilder) Apply(ctx context.Context, b multiMediaBuilder) error {
	return Media(&u.doc, u.caption...).Apply(ctx, b)
}

// PhotoExternal adds document attachment that will be downloaded by the Telegram servers.
func PhotoExternal(url string, caption ...StyledTextOption) *PhotoExternalBuilder {
	return &PhotoExternalBuilder{
		doc: tg.InputMediaPhotoExternal{
			URL: url,
		},
		caption: caption,
	}
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

// Apply implements MediaOption.
func (u *UploadedPhotoBuilder) Apply(ctx context.Context, b multiMediaBuilder) error {
	return Media(&u.photo, u.caption...).Apply(ctx, b)
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
