package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/tg"
)

// Manager is peer manager.
type Manager struct {
	api     *tg.Client
	storage Storage
	cache   Cache

	me *atomicUser

	logger *zap.Logger
	phone  *semaphore.Weighted
	sg     singleflight.Group
}

// NewManager creates new Manager.
func NewManager(api *tg.Client, opts Options) *Manager {
	opts.setDefaults()
	return &Manager{
		api:     api,
		storage: opts.Storage,
		cache:   opts.Cache,
		me:      new(atomicUser),
		logger:  opts.Logger,
		phone:   semaphore.NewWeighted(1),
		sg:      singleflight.Group{},
	}
}

// Init initializes Manager.
func (m *Manager) Init(ctx context.Context) error {
	_, err := m.Self(ctx)
	if err != nil {
		return errors.Wrap(err, "get self")
	}
	return nil
}

// SetChannelAccessHash implements updates.ChannelAccessHasher.
func (m *Manager) SetChannelAccessHash(userID, channelID, accessHash int64) error {
	myID, ok := m.myID()
	if !ok || myID != userID {
		return nil
	}
	return m.storage.Save(context.TODO(), Key{
		Prefix: channelPrefix,
		ID:     channelID,
	}, Value{
		AccessHash: accessHash,
	})
}

// GetChannelAccessHash implements updates.ChannelAccessHasher.
func (m *Manager) GetChannelAccessHash(userID, channelID int64) (accessHash int64, found bool, err error) {
	myID, ok := m.myID()
	if !ok || myID != userID {
		return 0, false, nil
	}
	v, found, err := m.storage.Find(context.TODO(), Key{
		Prefix: channelPrefix,
		ID:     channelID,
	})
	return v.AccessHash, found, err
}
