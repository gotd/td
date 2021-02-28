package message

import "github.com/gotd/td/tg"

// Video adds video attachment.
func Video(file tg.InputFileClass, mime string, caption ...StyledTextOption) *UploadedDocumentBuilder {
	// TODO(tdakkota): better MIME and attributes building.
	return UploadedDocument(file, caption...).
		MIME(mime)
}

// RoundVideo adds round video attachment.
func RoundVideo(file tg.InputFileClass, mime string, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return Video(file, mime, caption...).Attributes(&tg.DocumentAttributeVideo{
		RoundMessage: true,
	})
}

// GIF adds gif attachment.
func GIF(file tg.InputFileClass, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return UploadedDocument(file, caption...).NosoundVideo(true)
}
