package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/tg"
)

func TestConnectionStateString(t *testing.T) {
	for _, tt := range []struct {
		state    ConnectionState
		expected string
	}{
		{ConnectionStateConnecting, "connecting"},
		{ConnectionStateReady, "ready"},
		{ConnectionStateDisconnected, "disconnected"},
		{ConnectionState(-1), "unknown"},
	} {
		require.Equal(t, tt.expected, tt.state.String())
	}
}

func TestClient_notifyConnectionState(t *testing.T) {
	var states []ConnectionState
	c := Client{
		onConnectionState: func(s ConnectionState) {
			states = append(states, s)
		},
	}
	c.notifyConnectionState(ConnectionStateConnecting)
	c.notifyConnectionState(ConnectionStateReady)
	require.Equal(t, []ConnectionState{
		ConnectionStateConnecting,
		ConnectionStateReady,
	}, states)

	// Callback is optional.
	c = Client{}
	require.NotPanics(t, func() {
		c.notifyConnectionState(ConnectionStateReady)
	})
}

// TestClient_onSessionNoConnectionState ensures that the shared OnSession
// handler does not emit connection state events: it is used by pool/sub
// connections (including same-DC ones), so emitting here would produce
// spurious ready events for non-primary connections.
func TestClient_onSessionNoConnectionState(t *testing.T) {
	var states []ConnectionState
	c := &Client{
		log:               log.For(log.Nop),
		onConnectionState: func(s ConnectionState) { states = append(states, s) },
		session:           pool.NewSyncSession(pool.Session{DC: 2}),
	}
	c.init()

	// Simulate OnSession arriving from a same-DC pool/sub connection.
	require.NoError(t, c.onSession(tg.Config{ThisDC: 2}, mtproto.Session{}))
	require.Empty(t, states, "onSession must not emit connection state")
}
