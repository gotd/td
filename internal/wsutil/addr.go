package wsutil

// Addr is a simple net.Addr implementation.
type Addr string

// Network implements net.Addr.
// Always returns "websocket".
func (addr Addr) Network() string {
	return "websocket"
}

// String implements net.Addr and fmt.Stringer.
// Returns Addr value.
func (addr Addr) String() string {
	return string(addr)
}
