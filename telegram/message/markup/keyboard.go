package markup

import (
	"github.com/nnqq/td/tg"
)

// ReplyKeyboardMarkupBuilder is a keyboard markup builder.
type ReplyKeyboardMarkupBuilder struct {
	kb tg.ReplyKeyboardMarkup
}

// Resize sets flag to request clients to resize the keyboard vertically for
// optimal fit (e.g., make the keyboard smaller if there are just two rows of buttons).
// If not set, the custom keyboard is always of the same height as the app's standard keyboard.
func (b *ReplyKeyboardMarkupBuilder) Resize() *ReplyKeyboardMarkupBuilder {
	b.kb.Resize = true
	return b
}

// SingleUse sets flag to request clients to hide the keyboard as soon as it's been used.
// The keyboard will still be available, but clients will automatically display the usual letter-keyboard
// in the chat – the user can press a special button in the input field to see the custom keyboard again.
func (b *ReplyKeyboardMarkupBuilder) SingleUse() *ReplyKeyboardMarkupBuilder {
	b.kb.SingleUse = true
	return b
}

// Selective sets flag to show the keyboard to specific users only.
// Targets:
// 	1) users that are @mentioned in the text of the Message object;
// 	2) if the bot's message is a reply (has reply_to_message_id), sender of the original message.
//
// Example: A user requests to change the bot‘s language, bot replies to the request
// with a keyboard to select the new language.
// Other users in the group don’t see the keyboard.
func (b *ReplyKeyboardMarkupBuilder) Selective() *ReplyKeyboardMarkupBuilder {
	b.kb.Selective = true
	return b
}

// Build returns created keyboard.
func (b *ReplyKeyboardMarkupBuilder) Build(rows ...tg.KeyboardButtonRow,
) tg.ReplyMarkupClass {
	cp := b.kb
	cp.Rows = rows
	return &cp
}

// BuildKeyboard creates keyboard builder.
func BuildKeyboard() *ReplyKeyboardMarkupBuilder {
	return &ReplyKeyboardMarkupBuilder{}
}

// SingleRow creates keyboard with single row using given buttons.
func SingleRow(buttons ...tg.KeyboardButtonClass) tg.ReplyMarkupClass {
	return Keyboard(Row(buttons...))
}

// Keyboard creates keyboard using given rows.
func Keyboard(rows ...tg.KeyboardButtonRow) tg.ReplyMarkupClass {
	return BuildKeyboard().Build(rows...)
}
