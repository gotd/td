package peers

import (
	"github.com/gotd/log"
	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// Options is options of Manager
type Options struct {
	Storage Storage
	// Cache stores users, chats and channels, so that resolving a peer that
	// was already seen (e.g. in an update) makes no RPC. Defaults to
	// InmemoryCache, which keeps every peer for the lifetime of the Manager,
	// like TDLib and Telegram Desktop do; set NoopCache{} to fetch every time.
	Cache  Cache
	Logger log.Logger
	// Clock is used for the expiry of cached usernames (UsernameTTL).
	// Defaults to clock.System.
	Clock clock.Clock
}

func (o *Options) setDefaults() {
	if o.Storage == nil {
		o.Storage = &InmemoryStorage{}
	}
	if o.Cache == nil {
		o.Cache = &InmemoryCache{}
	}
	if o.Clock == nil {
		o.Clock = clock.System
	}
	if o.Logger == nil {
		o.Logger = log.Nop
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
		logger:  log.For(o.Logger),
		sg:      singleflight.Group{},
		clock:   o.Clock,

		usernames: newUsernameCache(),
		needUpdate: peerIDSet{
			m: make(map[constant.TDLibPeerID]struct{}),
		},
		needUpdateFull: peerIDSet{
			m: make(map[constant.TDLibPeerID]struct{}),
		},
	}
}
