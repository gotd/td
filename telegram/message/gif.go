package message

import (
	"context"

	"github.com/gotd/td/tg"
)

// GIF adds gif attachment.
func GIF(file tg.InputFileClass, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return UploadedDocument(file, caption...).
		Attributes(&tg.DocumentAttributeAnimated{}).
		MIME("image/gif")
}

// GIF sends gif.
func (b *Builder) GIF(ctx context.Context, file tg.InputFileClass, caption ...StyledTextOption) error {
	return b.Media(ctx, GIF(file, caption...))
}
