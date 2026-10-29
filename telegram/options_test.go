package telegram

import (
	"testing"
	"time"

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

func TestOptionsPingForwarding(t *testing.T) {
	client := NewClient(1, "hash", Options{
		PingInterval:      30 * time.Second,
		PingTimeout:       20 * time.Second,
		PingProgressAware: true,
		PingMaxWait:       2 * time.Minute,
	})
	require.Equal(t, 30*time.Second, client.opts.PingInterval)
	require.Equal(t, 20*time.Second, client.opts.PingTimeout)
	require.True(t, client.opts.PingProgressAware)
	require.Equal(t, 2*time.Minute, client.opts.PingMaxWait)
}
