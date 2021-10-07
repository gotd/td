package message

import (
	"context"
	"time"

	"github.com/nnqq/td/tg"
)

// Audio creates new AudioDocumentBuilder to create audio attachment.
func (u *UploadedDocumentBuilder) Audio() *AudioDocumentBuilder {
	b := u
	if u.doc.MimeType == "" {
		b = u.MIME(DefaultAudioMIME)
	}
	return &AudioDocumentBuilder{
		doc:  b,
		attr: tg.DocumentAttributeAudio{},
	}
}

// Voice creates new AudioDocumentBuilder to create voice attachment.
func (u *UploadedDocumentBuilder) Voice() *AudioDocumentBuilder {
	return u.MIME(DefaultVoiceMIME).Audio().Voice()
}

// AudioDocumentBuilder is an Audio media option.
type AudioDocumentBuilder struct {
	doc  *UploadedDocumentBuilder
	attr tg.DocumentAttributeAudio
}

// Voice sets flag to mark this audio as voice message.
func (u *AudioDocumentBuilder) Voice() *AudioDocumentBuilder {
	u.attr.Voice = true
	return u
}

// Duration sets duration of audio file.
func (u *AudioDocumentBuilder) Duration(duration time.Duration) *AudioDocumentBuilder {
	return u.DurationSeconds(int(duration.Seconds()))
}

// DurationSeconds sets duration in seconds.
func (u *AudioDocumentBuilder) DurationSeconds(duration int) *AudioDocumentBuilder {
	u.attr.Duration = duration
	return u
}

// Title sets name of song.
func (u *AudioDocumentBuilder) Title(title string) *AudioDocumentBuilder {
	u.attr.Title = title
	return u
}

// Performer sets performer.
func (u *AudioDocumentBuilder) Performer(performer string) *AudioDocumentBuilder {
	u.attr.Performer = performer
	return u
}

// Waveform sets waveform representation of the voice message.
func (u *AudioDocumentBuilder) Waveform(waveform []byte) *AudioDocumentBuilder {
	u.attr.Waveform = waveform
	return u
}

// apply implements MediaOption.
func (u *AudioDocumentBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return u.doc.Attributes(&u.attr).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *AudioDocumentBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return u.doc.Attributes(&u.attr).applyMulti(ctx, b)
}

// Audio adds audio attachment.
func Audio(file tg.InputFileClass, caption ...StyledTextOption) *AudioDocumentBuilder {
	return UploadedDocument(file, caption...).Audio()
}

// Voice adds voice attachment.
func Voice(file tg.InputFileClass) *AudioDocumentBuilder {
	return UploadedDocument(file).Voice()
}

// Audio sends audio file.
func (b *Builder) Audio(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, Audio(file, caption...))
}

// Voice sends voice message.
func (b *Builder) Voice(ctx context.Context, file tg.InputFileClass) (tg.UpdatesClass, error) {
	return b.Media(ctx, Voice(file))
}
