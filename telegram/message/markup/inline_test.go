package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestInlineRow(t *testing.T) {
	a := require.New(t)

	buttons := []tg.KeyboardButtonClass{
		Button("gotd"),
		URL("Google!", "https://google.com?q=gotd"),
		RequestPhone("phone"),
		RequestGeoLocation("geo"),
		SwitchInline("inline", "query", true),
		Game("game"),
		Buy("buy"),
		InputURLAuth(false, "text", "fwdText", "url", &tg.InputUserSelf{}),
		RequestPoll("poll", true),
		InputUserProfile("me", &tg.InputUserSelf{}),
		WebView("demo", "https://webappcontent.telegram.org/demo"),
	}

	v, ok := InlineRow(buttons...).(*tg.ReplyInlineMarkup)
	a.True(ok)
	a.Len(v.Rows, 1)
	row := v.Rows[0]

	a.Len(row.Buttons, len(buttons))
	for i, b := range buttons {
		a.Equal(b, row.Buttons[i])
	}
}
