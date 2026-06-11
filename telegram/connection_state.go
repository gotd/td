package telegram

// ConnectionState represents the state of the primary connection.
//
// See Options.OnConnectionState.
type ConnectionState int

const (
	// ConnectionStateConnecting means that client is establishing the primary
	// connection (initial connect or reconnect after failure).
	ConnectionStateConnecting ConnectionState = iota
	// ConnectionStateReady means that the primary connection is initialized
	// and ready to send requests.
	ConnectionStateReady
	// ConnectionStateDisconnected means that the primary connection is dead.
	// Client will reconnect automatically, transitioning to
	// ConnectionStateConnecting.
	ConnectionStateDisconnected
)

// String implements fmt.Stringer.
func (s ConnectionState) String() string {
	switch s {
	case ConnectionStateConnecting:
		return "connecting"
	case ConnectionStateReady:
		return "ready"
	case ConnectionStateDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// notifyConnectionState calls OnConnectionState callback, if set.
func (c *Client) notifyConnectionState(s ConnectionState) {
	if c.onConnectionState != nil {
		c.onConnectionState(s)
	}
}
