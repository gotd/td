package pool

import "github.com/gotd/log"

// DCOptions is a Telegram data center connections pool options.
type DCOptions struct {
	// Logger is the structured logger. No logs by default.
	Logger log.Logger
	// MTProto options for connections.
	// Opened connection limit to the DC.
	MaxOpenConnections int64
	// RetryOnWriteFailed retries a request whose transport send failed on
	// another connection instead of returning the write error to the caller.
	//
	// Disabled by default: a caller that reacts to the write error itself —
	// rotating a proxy or an endpoint, for example — needs to keep seeing it.
	RetryOnWriteFailed bool
}

func (d *DCOptions) setDefaults() {
	if d.Logger == nil {
		d.Logger = log.Nop
	}
	// It's okay to use zero value for MaxOpenConnections.
}
