package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
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
