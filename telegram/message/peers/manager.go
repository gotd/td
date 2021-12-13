package peers

import (
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/tg"
)

// Manager is peer manager.
type Manager struct {
	api     *tg.Client
	storage Storage
	me      *atomicUser

	phone *semaphore.Weighted
	sg    singleflight.Group
}

// NewManager creates new Manager.
func NewManager(api *tg.Client, storage Storage) *Manager {
	return &Manager{
		api:     api,
		storage: storage,
		me:      new(atomicUser),
		phone:   semaphore.NewWeighted(1),
		sg:      singleflight.Group{},
	}
}

