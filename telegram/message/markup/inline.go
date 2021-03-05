package markup

import "github.com/gotd/td/tg"

// InlineRow creates inline keyboard with single row using given buttons.
func InlineRow(button tg.KeyboardButtonClass, buttons ...tg.KeyboardButtonClass) tg.ReplyMarkupClass {
	return InlineKeyboard(Row(button, buttons...))
}

// InlineKeyboard creates inline keyboard using given rows.
func InlineKeyboard(row tg.KeyboardButtonRow, rows ...tg.KeyboardButtonRow) tg.ReplyMarkupClass {
	return &tg.ReplyInlineMarkup{
		Rows: append([]tg.KeyboardButtonRow{row}, rows...),
	}
}

