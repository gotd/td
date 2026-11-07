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

// ConnKind is the kind of client connection.
type ConnKind int

const (
	// ConnKindPrimary is the primary connection of Client.Run.
	ConnKindPrimary ConnKind = iota
	// ConnKindData is a connection of a pool: Client.Pool, Client.DC,
	// Client.MediaOnly, Client.FileDC, or a FILE_MIGRATE connection.
	ConnKindData
	// ConnKindCDN is a connection to a CDN DC.
	ConnKindCDN
)

// String implements fmt.Stringer.
func (k ConnKind) String() string {
	switch k {
	case ConnKindPrimary:
		return "primary"
	case ConnKindData:
		return "data"
	case ConnKindCDN:
		return "cdn"
	default:
		return "unknown"
	}
}

// ConnDeadEvent describes a dead connection.
//
// See Options.OnConnDead.
type ConnDeadEvent struct {
	// DC is the ID of the DC of the connection.
	DC int
	// Kind is the kind of the connection.
	Kind ConnKind
	// Err is the reason.
	Err error
}

// ConnectionStateEvent describes a state change of the primary connection.
//
// See Options.OnConnectionStateEvent.
type ConnectionStateEvent struct {
	// DC is the ID of the DC of the primary connection.
	DC int
	// State is the new state.
	State ConnectionState
}

// notifyConnectionState calls OnConnectionState and OnConnectionStateEvent
// callbacks, if set.
func (c *Client) notifyConnectionState(s ConnectionState) {
	if c.onConnectionState != nil {
		c.onConnectionState(s)
	}
	if c.onConnectionStateEvent != nil {
		e := ConnectionStateEvent{State: s}
		if c.session != nil {
			e.DC = c.session.Load().DC
		}
		c.onConnectionStateEvent(e)
	}
}

// notifyConnDead calls OnDead and OnConnDead callbacks, if set.
func (c *Client) notifyConnDead(e ConnDeadEvent) {
	if c.onDead != nil {
		c.onDead(e.Err)
	}
	if c.onConnDead != nil {
		c.onConnDead(e)
	}
}
