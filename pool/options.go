package pool

import "github.com/gotd/log"

// DCOptions is a Telegram data center connections pool options.
type DCOptions struct {
	// Logger is the structured logger. No logs by default.
	Logger log.Logger
	// MTProto options for connections.
	// Opened connection limit to the DC.
	MaxOpenConnections int64
}

func (d *DCOptions) setDefaults() {
	if d.Logger == nil {
		d.Logger = log.Nop
	}
	// It's okay to use zero value for MaxOpenConnections.
}
