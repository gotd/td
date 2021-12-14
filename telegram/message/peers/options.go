package peers

import "go.uber.org/zap"

// Options is options of Manager
type Options struct {
	Storage Storage
	Logger  *zap.Logger
}

func (o *Options) setDefaults() {
	if o.Storage == nil {
		// TODO(tdakkota): set inmemory
	}
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
}
