package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func testInlineButtons() []tg.KeyboardInlineButton {
	return []tg.KeyboardInlineButton{
		URL("Google!", "https://google.com?q=gotd"),
		Callback("callback", []byte("data")),
		SwitchInline("inline", "query", true),
		Game("game"),
		Buy("buy"),
		InputURLAuth(false, "text", "fwdText", "url", &tg.InputUserSelf{}),
		URLAuth("text", "url", 1, "fwd"),
		InputUserProfile("me", &tg.InputUserSelf{}),
		UserProfile("BotFather", 93372553),
		WebView("demo", "https://webappcontent.telegram.org/demo"),
		Copy("copy", "text to copy"),
		Disabled("disabled"),
		InlineButton("generic", &tg.InlineButtonTypeURL{URL: "https://gotd.dev"}),
	}
}

func TestInlineRow(t *testing.T) {
	a := require.New(t)

	buttons := testInlineButtons()
	v, ok := InlineRow(buttons...).(*tg.ReplyInlineMarkup)
	a.True(ok)
	a.Len(v.Rows, 1)
	row := v.Rows[0]

	a.Len(row.Buttons, len(buttons))
	for i, b := range buttons {
		a.Equal(b, row.Buttons[i])
	}
}

func TestInlineKeyboardEncode(t *testing.T) {
	a := require.New(t)

	buttons := testInlineButtons()
	m, ok := InlineKeyboard(
		InlineButtonRow(buttons[:3]...),
		InlineButtonRow(buttons[3:]...),
	).(*tg.ReplyInlineMarkup)
	a.True(ok)

	var buf bin.Buffer
	a.NoError(m.Encode(&buf))

	var decoded tg.ReplyInlineMarkup
	a.NoError(decoded.Decode(&buf))
	a.Len(decoded.Rows, 2)
	a.Len(decoded.Rows[0].Buttons, 3)
	a.Len(decoded.Rows[1].Buttons, len(buttons)-3)
	a.Equal(&tg.InlineButtonTypeURL{URL: "https://google.com?q=gotd"}, decoded.Rows[0].Buttons[0].Type)
	a.Equal("copy", decoded.Rows[1].Buttons[7].Text)
}
