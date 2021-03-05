package message

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Upload creates new UploadBuilder to upload and send attachments.
// Given option will be called only once, even if you call upload functions .
func (b *RequestBuilder) Upload(upd UploadOption) *UploadBuilder {
	return &UploadBuilder{
		builder: *b,
		option:  upd,
	}
}

// UploadBuilder is a attachment uploading helper.
type UploadBuilder struct {
	builder RequestBuilder
	option  UploadOption
}

func (u *UploadBuilder) next() *UploadBuilder {
	return u
}

// Silent sets flag to send this message silently (no notifications for the receivers).
func (u *UploadBuilder) Silent() *UploadBuilder {
	r := u.next()
	r.builder.silent = true
	return r
}

// Background sets flag to send this message as background message.
func (u *UploadBuilder) Background() *UploadBuilder {
	r := u.next()
	r.builder.background = true
	return r
}

// ClearDraft sets flag to clear the draft field.
func (u *UploadBuilder) ClearDraft() *UploadBuilder {
	r := u.next()
	r.builder.clearDraft = true
	return r
}

// Reply sets message ID to reply.
func (u *UploadBuilder) Reply(id int) *UploadBuilder {
	r := u.next()
	r.builder.replyToMsgID = id
	return r
}

// ReplyMsg sets message to reply.
func (u *UploadBuilder) ReplyMsg(msg tg.MessageClass) *UploadBuilder {
	return u.Reply(msg.GetID())
}

// Schedule sets scheduled message date for scheduled messages.
func (u *UploadBuilder) Schedule(date time.Time) *UploadBuilder {
	r := u.next()
	r.builder.scheduleDate = int(date.Unix())
	return r
}

// NoWebpage sets flag to disable generation of the webpage preview.
func (u *UploadBuilder) NoWebpage() *UploadBuilder {
	r := u.next()
	r.builder.noWebpage = true
	return r
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (u *UploadBuilder) Markup(markup tg.ReplyMarkupClass) *UploadBuilder {
	r := u.next()
	r.builder.replyMarkup = markup
	return r
}

func (u *UploadBuilder) file(ctx context.Context) (tg.InputFileClass, error) {
	return u.option.apply(ctx, uploadBuilder{upload: u.builder.sender.uploader})
}

// AsInputFile uploads and returns uploaded file entity.
func (u *UploadBuilder) AsInputFile(ctx context.Context) (tg.InputFileClass, error) {
	return u.file(ctx)
}

func (u *UploadBuilder) send(
	ctx context.Context,
	using func(ctx context.Context, file tg.InputFileClass, caption ...StyledTextOption) error,
	caption []StyledTextOption,
) error {
	f, err := u.file(ctx)
	if err != nil {
		return xerrors.Errorf("upload: %w", err)
	}

	return using(ctx, f, caption...)
}

// Photo uploads and sends photo.
func (u *UploadBuilder) Photo(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.UploadedPhoto, caption)
}

// Audio uploads and sends audio file.
func (u *UploadBuilder) Audio(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.Audio, caption)
}

// Voice uploads and sends voice message.
func (u *UploadBuilder) Voice(ctx context.Context) error {
	f, err := u.file(ctx)
	if err != nil {
		return xerrors.Errorf("upload: %w", err)
	}

	return u.builder.Voice(ctx, f)
}

// Video uploads and sends video.
func (u *UploadBuilder) Video(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.Video, caption)
}

// RoundVideo uploads and sends round video.
func (u *UploadBuilder) RoundVideo(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.RoundVideo, caption)
}

// GIF uploads and sends gif file.
func (u *UploadBuilder) GIF(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.GIF, caption)
}

// Sticker uploads and sends sticker.
func (u *UploadBuilder) Sticker(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.UploadedSticker, caption)
}

// File uploads and sends plain file.
func (u *UploadBuilder) File(ctx context.Context, caption ...StyledTextOption) error {
	return u.send(ctx, u.builder.File, caption)
}
