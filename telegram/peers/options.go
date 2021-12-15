package peers

import "go.uber.org/zap"

// Options is options of Manager
type Options struct {
	Storage Storage
	Cache   Cache
	Logger  *zap.Logger
}

func (o *Options) setDefaults() {
	if o.Storage == nil {
		o.Storage = &InmemoryStorage{}
	}
	if o.Cache == nil {
		o.Cache = NoopCache{}
	}
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
}
