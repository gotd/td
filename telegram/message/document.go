package message

import (
	"context"
	"encoding/hex"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// DocumentBuilder is a Document media option.
type DocumentBuilder struct {
	doc     tg.InputMediaDocument
	caption []StyledTextOption
}

// TTL sets time to live of self-destructing document.
func (u *DocumentBuilder) TTL(ttl time.Duration) *DocumentBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing document.
func (u *DocumentBuilder) TTLSeconds(ttl int) *DocumentBuilder {
	u.doc.TTLSeconds = ttl
	return u
}

// Query sets query field of InputMediaDocument.
func (u *DocumentBuilder) Query(query string) *DocumentBuilder {
	u.doc.Query = query
	return u
}

// apply implements MediaOption.
func (u *DocumentBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.doc, u.caption...).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *DocumentBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return u.apply(ctx, b)
}

// Document adds document attachment.
func Document(doc FileLocation, caption ...StyledTextOption) *DocumentBuilder {
	v := new(tg.InputDocument)
	v.FillFrom(doc)

	return &DocumentBuilder{
		doc: tg.InputMediaDocument{
			ID: v,
		},
		caption: caption,
	}
}

// Document sends document.
func (b *Builder) Document(
	ctx context.Context, file FileLocation, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, Document(file, caption...))
}

// SearchDocumentBuilder is a Document media option which uses messages.getDocumentByHash
// to find document.
//
// See https://core.telegram.org/method/messages.getDocumentByHash.
//
// See https://core.telegram.org/api/files#re-using-pre-uploaded-files.
type SearchDocumentBuilder struct {
	hash    []byte
	size    int
	mime    string
	builder *DocumentBuilder
}

// TTL sets time to live of self-destructing document.
func (u *SearchDocumentBuilder) TTL(ttl time.Duration) *SearchDocumentBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing document.
func (u *SearchDocumentBuilder) TTLSeconds(ttl int) *SearchDocumentBuilder {
	u.builder.doc.TTLSeconds = ttl
	return u
}

// Query sets query field of InputMediaDocument.
func (u *SearchDocumentBuilder) Query(query string) *SearchDocumentBuilder {
	u.builder.doc.Query = query
	return u
}

// apply implements MediaOption.
func (u *SearchDocumentBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	result, err := b.sender.getDocumentByHash(ctx, &tg.MessagesGetDocumentByHashRequest{
		SHA256:   u.hash,
		Size:     u.size,
		MimeType: u.mime,
	})
	if err != nil {
		return xerrors.Errorf("find document: %w", err)
	}

	doc, ok := result.AsNotEmpty()
	if !ok {
		return xerrors.Errorf("document with hash %q not found", hex.EncodeToString(u.hash))
	}

	v := new(tg.InputDocument)
	v.FillFrom(doc)
	u.builder.doc.ID = v
	return Media(&u.builder.doc, u.builder.caption...).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *SearchDocumentBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return u.apply(ctx, b)
}

// DocumentByHash finds document by hash and adds as attachment.
//
// See https://core.telegram.org/method/messages.getDocumentByHash.
//
// See https://core.telegram.org/api/files#re-using-pre-uploaded-files.
func DocumentByHash(
	hash []byte, size int, mime string,
	caption ...StyledTextOption,
) *SearchDocumentBuilder {
	return &SearchDocumentBuilder{
		hash: hash,
		size: size,
		mime: mime,
		builder: &DocumentBuilder{
			caption: caption,
		},
	}
}

// DocumentByHash finds document by hash and sends as attachment.
func (b *Builder) DocumentByHash(
	ctx context.Context, hash []byte, size int, mime string,
	caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, DocumentByHash(hash, size, mime, caption...))
}

// DocumentExternalBuilder is a DocumentExternal media option.
type DocumentExternalBuilder struct {
	doc     tg.InputMediaDocumentExternal
	caption []StyledTextOption
}

// TTL sets time to live of self-destructing document.
func (u *DocumentExternalBuilder) TTL(ttl time.Duration) *DocumentExternalBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing document.
func (u *DocumentExternalBuilder) TTLSeconds(ttl int) *DocumentExternalBuilder {
	u.doc.TTLSeconds = ttl
	return u
}

// apply implements MediaOption.
func (u *DocumentExternalBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.doc, u.caption...).apply(ctx, b)
}

// DocumentExternal adds document attachment that will be downloaded by the Telegram servers.
func DocumentExternal(url string, caption ...StyledTextOption) *DocumentExternalBuilder {
	return &DocumentExternalBuilder{
		doc: tg.InputMediaDocumentExternal{
			URL: url,
		},
		caption: caption,
	}
}

// DocumentExternal sends document attachment that will be downloaded by the Telegram servers.
func (b *Builder) DocumentExternal(ctx context.Context, url string, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return b.Media(ctx, DocumentExternal(url, caption...))
}

// UploadedDocumentBuilder is a UploadedDocument media option.
type UploadedDocumentBuilder struct {
	doc     tg.InputMediaUploadedDocument
	caption []StyledTextOption
}

// NosoundVideo sets flag that the specified document is a video file with no audio tracks
// (a GIF animation (even as MPEG4), for example).
func (u *UploadedDocumentBuilder) NosoundVideo(v bool) *UploadedDocumentBuilder {
	u.doc.NosoundVideo = v
	return u
}

// ForceFile sets flag to force the media file to be uploaded as document.
func (u *UploadedDocumentBuilder) ForceFile(v bool) *UploadedDocumentBuilder {
	u.doc.ForceFile = v
	return u
}

// Thumb sets thumbnail of the document, uploaded as for the file.
func (u *UploadedDocumentBuilder) Thumb(file tg.InputFileClass) *UploadedDocumentBuilder {
	u.doc.Thumb = file
	return u
}

// MIME sets MIME type of document.
func (u *UploadedDocumentBuilder) MIME(mime string) *UploadedDocumentBuilder {
	u.doc.MimeType = mime
	return u
}

// Attributes adds given attributes to the document.
// Attribute specify the type of the document (video, audio, voice, sticker, etc.).
func (u *UploadedDocumentBuilder) Attributes(attrs ...tg.DocumentAttributeClass) *UploadedDocumentBuilder {
	u.doc.Attributes = append(u.doc.Attributes, attrs...)
	return u
}

// Filename sets name of uploaded file.
func (u *UploadedDocumentBuilder) Filename(name string) *UploadedDocumentBuilder {
	return u.Attributes(&tg.DocumentAttributeFilename{
		FileName: name,
	})
}

// HasStickers sets flag that document attachment has stickers.
func (u *UploadedDocumentBuilder) HasStickers() *UploadedDocumentBuilder {
	return u.Attributes(&tg.DocumentAttributeHasStickers{})
}

// Stickers adds attached mask stickers.
func (u *UploadedDocumentBuilder) Stickers(stickers ...FileLocation) *UploadedDocumentBuilder {
	u.doc.Stickers = append(u.doc.Stickers, inputDocuments(stickers...)...)
	return u
}

// TTL sets time to live of self-destructing document.
func (u *UploadedDocumentBuilder) TTL(ttl time.Duration) *UploadedDocumentBuilder {
	return u.TTLSeconds(int(ttl.Seconds()))
}

// TTLSeconds sets time to live in seconds of self-destructing document.
func (u *UploadedDocumentBuilder) TTLSeconds(ttl int) *UploadedDocumentBuilder {
	u.doc.TTLSeconds = ttl
	return u
}

// apply implements MediaOption.
func (u *UploadedDocumentBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.doc, u.caption...).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *UploadedDocumentBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	m, err := b.sender.uploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  b.peer,
		Media: &u.doc,
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

// UploadedDocument adds document attachment.
func UploadedDocument(file tg.InputFileClass, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return &UploadedDocumentBuilder{
		doc: tg.InputMediaUploadedDocument{
			File: file,
		},
		caption: caption,
	}
}

// File adds document attachment and forces it to be used as plain file, not media.
func File(file tg.InputFileClass, caption ...StyledTextOption) *UploadedDocumentBuilder {
	return UploadedDocument(file, caption...).ForceFile(true)
}

// File sends uploaded file as document and forces it to be used as plain file, not media.
func (b *Builder) File(ctx context.Context, file tg.InputFileClass, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return b.Media(ctx, File(file, caption...))
}
