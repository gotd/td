package inline

import "github.com/nnqq/td/tg"

// DocumentResultBuilder is document result option builder.
type DocumentResultBuilder struct {
	result *tg.InputBotInlineResultDocument
	msg    MessageOption
}

func (b *DocumentResultBuilder) apply(r *resultPageBuilder) error {
	m, err := b.msg.apply()
	if err != nil {
		return err
	}

	t := tg.InputBotInlineResultDocument{
		ID:          b.result.ID,
		Type:        b.result.Type,
		Title:       b.result.Title,
		Description: b.result.Description,
		Document:    b.result.Document,
	}
	if t.ID == "" {
		t.ID, err = r.generateID()
		if err != nil {
			return err
		}
	}

	t.SendMessage = m
	r.results = append(r.results, &t)
	return nil
}

// ID sets ID of result.
// Should not be empty, so if id is not provided, random will be used.
func (b *DocumentResultBuilder) ID(id string) *DocumentResultBuilder {
	b.result.ID = id
	return b
}

// Title sets Result description.
func (b *DocumentResultBuilder) Title(title string) *DocumentResultBuilder {
	b.result.SetTitle(title)
	return b
}

// Description sets Result description.
func (b *DocumentResultBuilder) Description(description string) *DocumentResultBuilder {
	b.result.SetDescription(description)
	return b
}

// Document creates document result option builder.
func Document(doc tg.InputDocumentClass, typ string, msg MessageOption) *DocumentResultBuilder {
	return &DocumentResultBuilder{
		result: &tg.InputBotInlineResultDocument{
			Type:     typ,
			Document: doc,
		},
		msg: msg,
	}
}

// Video creates video result option builder.
func Video(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, VideoType, msg)
}

// Audio creates audio result option builder.
func Audio(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, AudioType, msg)
}

// File creates document result option builder.
func File(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, DocumentType, msg)
}

// GIF creates gif result option builder.
func GIF(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, GIFType, msg)
}

// MPEG4GIF creates mpeg4gif result option builder.
func MPEG4GIF(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, MPEG4GIFType, msg)
}

// Voice creates voice result option builder.
func Voice(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, VoiceType, msg)
}

// Sticker creates sticker result option builder.
func Sticker(doc tg.InputDocumentClass, msg MessageOption) *DocumentResultBuilder {
	return Document(doc, StickerType, msg)
}
