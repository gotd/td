package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestSingleRow(t *testing.T) {
	a := require.New(t)

	buttons := []tg.KeyboardButtonClass{
		Button("gotd"),
		URL("Google!", "https://google.com?q=gotd"),
		RequestPhone("phone"),
		RequestGeoLocation("geo"),
		SwitchInline("inline", "query", true),
		Game("game"),
		Buy("buy"),
		URLAuth("text", "url", 1, "fwd"),
		RequestPoll("poll", true),
	}

	v, ok := SingleRow(buttons...).(*tg.ReplyKeyboardMarkup)
	a.True(ok)
	a.Len(v.Rows, 1)
	row := v.Rows[0]

	a.Len(row.Buttons, len(buttons))
	for i, b := range buttons {
		a.Equal(b, row.Buttons[i])
	}
}
