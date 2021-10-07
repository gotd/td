package inline

import (
	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/markup"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

// MessageTextBuilder is a builder of inline result text message.
type MessageTextBuilder struct {
	message *tg.InputBotInlineMessageText
	options []styling.StyledTextOption
}

func (b *MessageTextBuilder) apply() (tg.InputBotInlineMessageClass, error) {
	tb := entity.Builder{}
	if err := styling.Perform(&tb, b.options...); err != nil {
		return nil, err
	}
	msg, entities := tb.Complete()
	r := *b.message

	r.Message = msg
	r.Entities = entities
	return &r, nil
}

// MessageText creates new message text option builder.
func MessageText(msg string) *MessageTextBuilder {
	return MessageStyledText(styling.Plain(msg))
}

// MessageStyledText creates new message text option builder.
func MessageStyledText(texts ...styling.StyledTextOption) *MessageTextBuilder {
	return &MessageTextBuilder{
		message: &tg.InputBotInlineMessageText{},
		options: texts,
	}
}

// NoWebpage sets flag to disable generation of the webpage preview.
func (b *MessageTextBuilder) NoWebpage() *MessageTextBuilder {
	b.message.NoWebpage = true
	return b
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageTextBuilder) Markup(m tg.ReplyMarkupClass) *MessageTextBuilder {
	b.message.ReplyMarkup = m
	return b
}

// Row sets single row keyboard markup  for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageTextBuilder) Row(buttons ...tg.KeyboardButtonClass) *MessageTextBuilder {
	return b.Markup(markup.InlineRow(buttons...))
}
