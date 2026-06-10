package message

import (
	"context"

	"github.com/gotd/td/tg"
)

// WebPageBuilder is a WebPage media option, used to attach a link preview to a
// message with explicit options (forced large/small media, optional preview).
//
// See https://core.telegram.org/constructor/inputMediaWebPage.
type WebPageBuilder struct {
	webpage tg.InputMediaWebPage
	caption []StyledTextOption
}

// ForceLargeMedia sets flag to use a large media preview.
func (u *WebPageBuilder) ForceLargeMedia(v bool) *WebPageBuilder {
	u.webpage.ForceLargeMedia = v
	return u
}

// ForceSmallMedia sets flag to use a small media preview.
func (u *WebPageBuilder) ForceSmallMedia(v bool) *WebPageBuilder {
	u.webpage.ForceSmallMedia = v
	return u
}

// Optional sets flag to not emit a WEBPAGE_NOT_FOUND error if a preview cannot
// be generated for the URL (the message is sent without preview instead).
//
// Note: if the message text is also empty, a MESSAGE_EMPTY error is emitted.
func (u *WebPageBuilder) Optional(v bool) *WebPageBuilder {
	u.webpage.Optional = v
	return u
}

// apply implements MediaOption.
func (u *WebPageBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	return Media(&u.webpage, u.caption...).apply(ctx, b)
}

// WebPage adds a link preview for the given URL as media attachment.
//
// See https://core.telegram.org/constructor/inputMediaWebPage.
func WebPage(url string, caption ...StyledTextOption) *WebPageBuilder {
	return &WebPageBuilder{
		webpage: tg.InputMediaWebPage{
			URL: url,
		},
		caption: caption,
	}
}

// WebPage sends a message with a link preview for the given URL.
func (b *Builder) WebPage(
	ctx context.Context,
	url string, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return b.Media(ctx, WebPage(url, caption...))
}
