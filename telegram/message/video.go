package message

import (
	"context"
	"time"

	"github.com/nnqq/td/tg"
)

// Video creates new VideoDocumentBuilder to create video attachment.
func (u *UploadedDocumentBuilder) Video() *VideoDocumentBuilder {
	b := u
	if u.doc.MimeType == "" {
		b = u.MIME(DefaultVideoMIME)
	}
	return &VideoDocumentBuilder{
		doc:  b,
		attr: tg.DocumentAttributeVideo{},
	}
}

// RoundVideo creates new VideoDocumentBuilder to create round video attachment.
func (u *UploadedDocumentBuilder) RoundVideo() *VideoDocumentBuilder {
	return u.Video().Round()
}

// VideoDocumentBuilder is a Video media option.
type VideoDocumentBuilder struct {
	doc  *UploadedDocumentBuilder
	attr tg.DocumentAttributeVideo
}

// Round sets flag to mark this video as round.
func (u *VideoDocumentBuilder) Round() *VideoDocumentBuilder {
	u.attr.RoundMessage = true
	return u
}

// SupportsStreaming sets flag to mark this video as which supports streaming.
func (u *VideoDocumentBuilder) SupportsStreaming() *VideoDocumentBuilder {
	u.attr.SupportsStreaming = true
	return u
}

// Resolution sets resolution of this video.
func (u *VideoDocumentBuilder) Resolution(w, h int) *VideoDocumentBuilder {
	u.attr.W = w
	u.attr.H = h
	return u
}

// Duration sets duration of video file.
func (u *VideoDocumentBuilder) Duration(duration time.Duration) *VideoDocumentBuilder {
	return u.DurationSeconds(int(duration.Seconds()))
}

// DurationSeconds sets duration in seconds.
func (u *VideoDocumentBuilder) DurationSeconds(duration int) *VideoDocumentBuilder {
	u.attr.Duration = duration
	return u
}

// apply implements MediaOption.
func (u *VideoDocumentBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return u.doc.Attributes(&u.attr).apply(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (u *VideoDocumentBuilder) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return u.doc.Attributes(&u.attr).applyMulti(ctx, b)
}

// Video adds video attachment.
func Video(file tg.InputFileClass, caption ...StyledTextOption) *VideoDocumentBuilder {
	// TODO(tdakkota): better MIME and attributes building.
	return UploadedDocument(file, caption...).Video()
}

// RoundVideo adds round video attachment.
func RoundVideo(file tg.InputFileClass, caption ...StyledTextOption) *VideoDocumentBuilder {
	return UploadedDocument(file, caption...).RoundVideo()
}

// Video sends video.
func (b *Builder) Video(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, Video(file, caption...))
}

// RoundVideo sends round video.
func (b *Builder) RoundVideo(
	ctx context.Context,
	file tg.InputFileClass, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, RoundVideo(file, caption...))
}
