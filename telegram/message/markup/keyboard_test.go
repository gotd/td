package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestSingleRow(t *testing.T) {
	a := require.New(t)

	buttons := []tg.KeyboardButton{
		Button("gotd"),
		RequestPhone("phone"),
		RequestGeoLocation("geo"),
		RequestPoll("poll", true),
		SimpleWebView("demo", "https://webappcontent.telegram.org/demo"),
		RequestPeer("peer", 0, &tg.RequestPeerTypeUser{}),
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
