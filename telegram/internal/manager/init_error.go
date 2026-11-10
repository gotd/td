package manager

// InitError is returned by Conn.Run when the initConnection request
// (invokeWithLayer(initConnection(help.getConfig))) fails.
//
// It lets callers tell a rejected connection setup apart from errors of
// the transport, the key exchange or the setup callback.
type InitError struct {
	Err error
}

// Error implements error.
func (e *InitError) Error() string {
	return "initConnection: " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *InitError) Unwrap() error {
	return e.Err
}
