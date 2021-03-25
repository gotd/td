package inline

import "github.com/gotd/td/tg"

// DocumentResultBuilder is document result option builder.
type DocumentResultBuilder struct {
	result *tg.InputBotInlineResultDocument
	msg    MessageOption
}

// ID sets ID of result.
// Should not be empty, so if id is not provided, random will be used.
func (b *DocumentResultBuilder) ID(id string) *DocumentResultBuilder {
	b.result.ID = id
	return b
}

// Description sets Result description.
func (b *DocumentResultBuilder) Description(description string) *DocumentResultBuilder {
	b.result.Description = description
	return b
}

// Document creates document result option builder.
func Document(doc tg.InputDocumentClass, typ, title string, msg MessageOption) *DocumentResultBuilder {
	return &DocumentResultBuilder{
		result: &tg.InputBotInlineResultDocument{
			Type:     typ,
			Title:    title,
			Document: doc,
		},
		msg: msg,
	}
}

// Video creates video result option builder.
func Video(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, VideoType, title, msg)
}

// Animation creates animation result option builder.
func Animation(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, AnimationType, title, msg)
}

// Audio creates audio result option builder.
func Audio(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, AudioType, title, msg)
}

// File creates document result option builder.
func File(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, DocumentType, title, msg)
}

// GIF creates gif result option builder.
func GIF(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, GIFType, title, msg)
}

// MPEG4GIF creates mpeg4gif result option builder.
func MPEG4GIF(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, MPEG4GIFType, title, msg)
}

// Voice creates voice result option builder.
func Voice(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, VoiceType, title, msg)
}

// Sticker creates sticker result option builder.
func Sticker(doc tg.InputDocumentClass, title string, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, StickerType, title, msg)
}
