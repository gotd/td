package proto

const (
	// CodeAuthKeyNotFound means that specified auth key ID cannot be found by the DC.
	CodeAuthKeyNotFound = 404

	// CodeTransportFlood means that too many transport connections are
	// established to the same IP in a too short lapse of time, or if any
	// of the container/service message limits are reached.
	CodeTransportFlood = 429
)
