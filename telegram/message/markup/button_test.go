package markup

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestStyleOptions(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle(nil)
		a.True(style.Zero())
		a.False(style.BgPrimary)
		a.False(style.BgDanger)
		a.False(style.BgSuccess)
		_, ok := style.GetIcon()
		a.False(ok)
	})
	t.Run("BgPrimary", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle([]StyleOption{StyleBgPrimary()})
		a.True(style.BgPrimary)
		a.False(style.BgDanger)
		a.False(style.BgSuccess)
	})
	t.Run("BgDanger", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle([]StyleOption{StyleBgDanger()})
		a.False(style.BgPrimary)
		a.True(style.BgDanger)
		a.False(style.BgSuccess)
	})
	t.Run("BgSuccess", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle([]StyleOption{StyleBgSuccess()})
		a.False(style.BgPrimary)
		a.False(style.BgDanger)
		a.True(style.BgSuccess)
	})
	t.Run("Icon", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle([]StyleOption{StyleIcon(123456)})
		icon, ok := style.GetIcon()
		a.True(ok)
		a.Equal(int64(123456), icon)
	})
	t.Run("Combined", func(t *testing.T) {
		a := require.New(t)

		style := applyStyle([]StyleOption{
			StyleBgPrimary(),
			StyleBgDanger(),
			StyleBgSuccess(),
			StyleIcon(42),
		})
		a.True(style.BgPrimary)
		a.True(style.BgDanger)
		a.True(style.BgSuccess)
		icon, ok := style.GetIcon()
		a.True(ok)
		a.Equal(int64(42), icon)
	})
}

func TestButtonStyle(t *testing.T) {
	a := require.New(t)

	want := tg.KeyboardButtonStyle{BgDanger: true}
	want.SetIcon(777)

	options := []StyleOption{StyleBgDanger(), StyleIcon(777)}

	a.Equal(want, Button("text", options...).Style)
	a.Equal(want, URL("text", "url", options...).Style)
	a.Equal(want, Callback("text", []byte("data"), options...).Style)
	a.Equal(want, RequestPhone("text", options...).Style)
	a.Equal(want, RequestGeoLocation("text", options...).Style)
	a.Equal(want, SwitchInline("text", "query", true, options...).Style)
	a.Equal(want, Game("text", options...).Style)
	a.Equal(want, Buy("text", options...).Style)
	a.Equal(want, URLAuth("text", "url", 1, "fwd", options...).Style)
	a.Equal(want, RequestPoll("text", true, options...).Style)
	a.Equal(want, UserProfile("text", 1, options...).Style)
	a.Equal(want, WebView("text", "url", options...).Style)
	a.Equal(want, SimpleWebView("text", "url", options...).Style)
	a.Equal(want, RequestPeer("text", 0, &tg.RequestPeerTypeUser{}, options...).Style)
}

func TestButtonNoStyle(t *testing.T) {
	a := require.New(t)

	// Without options, the style must stay zero so SetFlags does not set the
	// optional Style flag during encoding.
	a.True(Button("text").Style.Zero())
	a.True(URL("text", "url").Style.Zero())
	a.True(Callback("text", []byte("data")).Style.Zero())
	a.True(RequestPeer("text", 0, &tg.RequestPeerTypeUser{}).Style.Zero())
}
