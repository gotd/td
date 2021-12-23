package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/gotd/td/tg"
)

// Manager is peer manager.
//
// NB: this package is completely experimental and still WIP.
type Manager struct {
	api     *tg.Client
	storage Storage
	cache   Cache

	me *atomicUser

	logger *zap.Logger
	sg     singleflight.Group
}

// Init initializes Manager.
func (m *Manager) Init(ctx context.Context) error {
	_, err := m.Self(ctx)
	if err != nil {
		return errors.Wrap(err, "get self")
	}
	return nil
}
