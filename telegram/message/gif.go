package message

import (
	"context"

	"github.com/nnqq/td/tg"
)

// GIF add attributes to create GIF attachment.
func (u *UploadedDocumentBuilder) GIF() *UploadedDocumentBuilder {
	return u.Attributes(&tg.DocumentAttributeAnimated{}).
		MIME(DefaultGifMIME)
}

// GIF adds gif attachment.
func GIF(file tg.InputFileClass, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return UploadedDocument(file, caption...).GIF()
}

// GIF sends gif.
func (b *Builder) GIF(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, GIF(file, caption...))
}
