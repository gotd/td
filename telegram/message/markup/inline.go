package markup

import "github.com/nnqq/td/tg"

// InlineRow creates inline keyboard with single row using given buttons.
func InlineRow(buttons ...tg.KeyboardButtonClass) tg.ReplyMarkupClass {
	return InlineKeyboard(Row(buttons...))
}

// InlineKeyboard creates inline keyboard using given rows.
func InlineKeyboard(rows ...tg.KeyboardButtonRow) tg.ReplyMarkupClass {
	return &tg.ReplyInlineMarkup{
		Rows: rows,
	}
}
