package message

import (
	"context"

	"github.com/nnqq/td/tg"
)

// UploadedSticker creates new UploadedStickerBuilder to create sticker attachment.
func (u *UploadedDocumentBuilder) UploadedSticker() *UploadedStickerBuilder {
	b := u
	if u.doc.MimeType == "" {
		b = u.MIME(DefaultStickerMIME)
	}
	return &UploadedStickerBuilder{
		doc: b,
		attr: tg.DocumentAttributeSticker{
			Stickerset: &tg.InputStickerSetEmpty{},
		},
	}
}

// UploadedSticker adds uploaded sticker attachment.
func UploadedSticker(file tg.InputFileClass, caption ...StyledTextOption) *UploadedStickerBuilder {
	return UploadedDocument(file, caption...).UploadedSticker()
}

// UploadedStickerBuilder is a uploaded sticker media option.
type UploadedStickerBuilder struct {
	doc  *UploadedDocumentBuilder
	attr tg.DocumentAttributeSticker
}

// Mask sets flag that is a mask sticker.
func (u *UploadedStickerBuilder) Mask(mask bool) *UploadedStickerBuilder {
	u.attr.Mask = mask
	return u
}

// Alt sets alternative emoji representation of sticker.
func (u *UploadedStickerBuilder) Alt(alt string) *UploadedStickerBuilder {
	u.attr.Alt = alt
	return u
}

// StickerSet sets associated sticker set.
func (u *UploadedStickerBuilder) StickerSet(stickerSet tg.InputStickerSetClass) *UploadedStickerBuilder {
	u.attr.Stickerset = stickerSet
	return u
}

// MaskCoords sets mask coordinates (if this is a mask sticker, attached to a photo).
func (u *UploadedStickerBuilder) MaskCoords(maskCoords tg.MaskCoords) *UploadedStickerBuilder {
	u.attr.MaskCoords = maskCoords
	return u
}

// apply implements MediaOption.
func (u *UploadedStickerBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return u.doc.Attributes(&u.attr).apply(ctx, b)
}

// UploadedSticker sends uploaded file as sticker.
func (b *Builder) UploadedSticker(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, UploadedSticker(file, caption...))
}
