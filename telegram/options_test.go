package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestOptionsLayer(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		opt := Options{}
		opt.setDefaults()
		require.Equal(t, tg.Layer, opt.Layer)
	})
	t.Run("Explicit", func(t *testing.T) {
		opt := Options{Layer: 42}
		opt.setDefaults()
		require.Equal(t, 42, opt.Layer)
	})
	t.Run("Client", func(t *testing.T) {
		require.Equal(t, 42, NewClient(1, "hash", Options{Layer: 42}).layer)
		require.Equal(t, tg.Layer, NewClient(1, "hash", Options{}).layer)
	})
}
