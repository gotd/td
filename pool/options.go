package pool

import "go.uber.org/zap"

// DCOptions is a Telegram data center connections pool options.
type DCOptions struct {
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// MTProto options for connections.
	// Opened connection limit to the DC.
	MaxOpenConnections int64
}

func (d *DCOptions) setDefaults() {
	if d.Logger == nil {
		d.Logger = zap.NewNop()
	}
	// It's okay to use zero value for MaxOpenConnections.
}
