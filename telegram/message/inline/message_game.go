package inline

import (
	"github.com/nnqq/td/telegram/message/markup"
	"github.com/nnqq/td/tg"
)

// MessageGameBuilder is a builder of inline result game message.
type MessageGameBuilder struct {
	message *tg.InputBotInlineMessageGame
}

// nolint:unparam
func (b *MessageGameBuilder) apply() (tg.InputBotInlineMessageClass, error) {
	r := *b.message
	return &r, nil
}

// MessageGame creates new message option builder.
func MessageGame() *MessageGameBuilder {
	return &MessageGameBuilder{
		message: &tg.InputBotInlineMessageGame{},
	}
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageGameBuilder) Markup(m tg.ReplyMarkupClass) *MessageGameBuilder {
	b.message.ReplyMarkup = m
	return b
}

// Row sets single row keyboard markup  for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageGameBuilder) Row(buttons ...tg.KeyboardButtonClass) *MessageGameBuilder {
	return b.Markup(markup.InlineRow(buttons...))
}
