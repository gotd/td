package peers

import (
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/tg"
)

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

// Build creates new Manager.
func (o Options) Build(api *tg.Client) *Manager {
	o.setDefaults()
	return &Manager{
		api:     api,
		storage: o.Storage,
		cache:   o.Cache,
		me:      new(atomicUser),
		logger:  o.Logger,
		sg:      singleflight.Group{},
	}
}
