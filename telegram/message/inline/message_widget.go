package inline

import (
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/widget"
	"github.com/gotd/td/tg"
)

// MessageWidgetBuilder is a builder of inline result HTML widget message.
type MessageWidgetBuilder struct {
	message *tg.InputBotInlineMessageWidget
	options []styling.StyledTextOption
}

func (b *MessageWidgetBuilder) apply() (tg.InputBotInlineMessageClass, error) {
	if err := widget.ValidateSize(b.message.Width, b.message.Height); err != nil {
		return nil, err
	}
	if err := widget.ValidateHTML(b.message.HTML); err != nil {
		return nil, err
	}
	tb := entity.Builder{}
	if err := styling.Perform(&tb, b.options...); err != nil {
		return nil, err
	}
	r := *b.message
	r.Message, r.Entities = tb.Complete()
	return &r, nil
}

// MessageWidget creates new HTML widget message option builder with optional
// caption. The html is checked with widget.ValidateHTML; the compressed size
// limit is enforced by the server (see widget.ValidateCompressed).
//
// Use it as content of a result, e.g. Article("Counter", MessageWidget(html)).
func MessageWidget(html string, caption ...styling.StyledTextOption) *MessageWidgetBuilder {
	return &MessageWidgetBuilder{
		message: &tg.InputBotInlineMessageWidget{HTML: html},
		options: caption,
	}
}

// Size sets widget width and height; zero width means full bubble width, zero
// height means that the height is not fixed.
func (b *MessageWidgetBuilder) Size(width, height int) *MessageWidgetBuilder {
	b.message.Width = width
	b.message.Height = height
	return b
}

// PublicData sets public data of the widget.
func (b *MessageWidgetBuilder) PublicData(data string) *MessageWidgetBuilder {
	b.message.PublicData = data
	return b
}

// ChatData sets data of the widget, which is specific to the chat.
func (b *MessageWidgetBuilder) ChatData(data string) *MessageWidgetBuilder {
	b.message.ChatData = data
	return b
}

// AllowQueries allows the widget to send queries to the bot
// (updateInlineBotWidgetQuery).
func (b *MessageWidgetBuilder) AllowQueries(allow bool) *MessageWidgetBuilder {
	b.message.AllowQueries = allow
	return b
}

// Shared makes the widget shared, so it can be sent again by its identifier.
func (b *MessageWidgetBuilder) Shared(shared bool) *MessageWidgetBuilder {
	b.message.SharedWidget = shared
	return b
}

// InvertMedia sets flag to show the caption above the widget.
func (b *MessageWidgetBuilder) InvertMedia(invert bool) *MessageWidgetBuilder {
	b.message.InvertMedia = invert
	return b
}

// Markup sets reply markup for sending bot buttons.
func (b *MessageWidgetBuilder) Markup(m tg.ReplyMarkupClass) *MessageWidgetBuilder {
	b.message.ReplyMarkup = m
	return b
}

// Row sets single row keyboard markup for sending bot buttons.
func (b *MessageWidgetBuilder) Row(buttons ...tg.KeyboardInlineButton) *MessageWidgetBuilder {
	return b.Markup(markup.InlineRow(buttons...))
}

// MessageSharedWidgetBuilder is a builder of inline result shared widget
// message.
type MessageSharedWidgetBuilder struct {
	message *tg.InputBotInlineMessageWidgetID
	options []styling.StyledTextOption
}

func (b *MessageSharedWidgetBuilder) apply() (tg.InputBotInlineMessageClass, error) {
	if b.message.WidgetID == 0 {
		return nil, widget.ErrInvalidWidgetID
	}
	tb := entity.Builder{}
	if err := styling.Perform(&tb, b.options...); err != nil {
		return nil, err
	}
	r := *b.message
	r.Message, r.Entities = tb.Complete()
	return &r, nil
}

// MessageSharedWidget creates new message option builder that sends an
// existing shared widget by its identifier, with optional caption.
func MessageSharedWidget(id int64, caption ...styling.StyledTextOption) *MessageSharedWidgetBuilder {
	return &MessageSharedWidgetBuilder{
		message: &tg.InputBotInlineMessageWidgetID{WidgetID: id},
		options: caption,
	}
}

// ChatData sets data of the widget, which is specific to the chat.
func (b *MessageSharedWidgetBuilder) ChatData(data string) *MessageSharedWidgetBuilder {
	b.message.ChatData = data
	return b
}

// InvertMedia sets flag to show the caption above the widget.
func (b *MessageSharedWidgetBuilder) InvertMedia(invert bool) *MessageSharedWidgetBuilder {
	b.message.InvertMedia = invert
	return b
}

// Markup sets reply markup for sending bot buttons.
func (b *MessageSharedWidgetBuilder) Markup(m tg.ReplyMarkupClass) *MessageSharedWidgetBuilder {
	b.message.ReplyMarkup = m
	return b
}

// Row sets single row keyboard markup for sending bot buttons.
func (b *MessageSharedWidgetBuilder) Row(buttons ...tg.KeyboardInlineButton) *MessageSharedWidgetBuilder {
	return b.Markup(markup.InlineRow(buttons...))
}
