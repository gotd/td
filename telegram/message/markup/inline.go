package markup

import "github.com/gotd/td/tg"

// InlineRow creates inline keyboard with single row using given buttons.
func InlineRow(buttons ...tg.KeyboardInlineButton) tg.ReplyMarkupClass {
	return InlineKeyboard(InlineButtonRow(buttons...))
}

// InlineKeyboard creates inline keyboard using given rows.
func InlineKeyboard(rows ...tg.KeyboardInlineButtonRow) tg.ReplyMarkupClass {
	return &tg.ReplyInlineMarkup{
		Rows: rows,
	}
}
