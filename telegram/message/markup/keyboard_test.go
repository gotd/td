package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func testKeyboardButtons() []tg.KeyboardButton {
	return []tg.KeyboardButton{
		Button("gotd"),
		RequestPhone("phone"),
		RequestGeoLocation("geo"),
		RequestPoll("poll", true),
		SimpleWebView("demo", "https://webappcontent.telegram.org/demo"),
		RequestPeer("peer", 0, &tg.RequestPeerTypeUser{}),
		KeyboardButton("generic", &tg.ButtonTypeDefault{}),
	}
}

func TestSingleRow(t *testing.T) {
	a := require.New(t)

	buttons := testKeyboardButtons()
	v, ok := SingleRow(buttons...).(*tg.ReplyKeyboardMarkup)
	a.True(ok)
	a.Len(v.Rows, 1)
	row := v.Rows[0]

	a.Len(row.Buttons, len(buttons))
	for i, b := range buttons {
		a.Equal(b, row.Buttons[i])
	}
}

func TestKeyboardEncode(t *testing.T) {
	a := require.New(t)

	buttons := testKeyboardButtons()
	m, ok := BuildKeyboard().Resize().Build(Row(buttons...)).(*tg.ReplyKeyboardMarkup)
	a.True(ok)

	var buf bin.Buffer
	a.NoError(m.Encode(&buf))

	var decoded tg.ReplyKeyboardMarkup
	a.NoError(decoded.Decode(&buf))
	a.True(decoded.Resize)
	a.Len(decoded.Rows, 1)
	a.Len(decoded.Rows[0].Buttons, len(buttons))
	a.Equal(&tg.ButtonTypeRequestPhone{}, decoded.Rows[0].Buttons[1].Type)
}
