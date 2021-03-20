package updates

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Handler will be called on received updates from Telegram.
type Handler interface {
	Handle(ctx context.Context, u *tg.Updates) error
	HandleShort(ctx context.Context, u *tg.UpdateShort) error
}

// Manager is a updates manager.
type Manager struct {
	handler Handler
	storage Storage

	raw *tg.Client
	log *zap.Logger
}

// NewManager creates new Manager.
func NewManager(handler Handler, storage Storage) *Manager {
	return &Manager{
		handler: handler,
		storage: storage,
		log:     zap.NewNop(),
	}
}

// SetRaw sets RPC client to use.
func (m *Manager) SetRaw(raw *tg.Client) {
	m.raw = raw
}

// WithLogger sets exchange flow logger.
func (m *Manager) WithLogger(log *zap.Logger) *Manager {
	m.log = log
	return m
}

// Sync checks if there is a updates gap and resolves it, if needed.
func (m *Manager) Sync(ctx context.Context) error {
	if err := m.gapCommon(ctx); err != nil {
		return err
	}

	if err := m.syncChannels(ctx); err != nil {
		return xerrors.Errorf("sync channels: %w", err)
	}
	return nil
}

func (m *Manager) gapCommon(ctx context.Context) error {
	// Syncing with remote state.
	remoteState, err := m.raw.UpdatesGetState(ctx)
	if err != nil {
		return xerrors.Errorf("get remote state: %w", err)
	}

	m.log.Info("Got common state",
		zap.Int("qts", remoteState.Qts),
		zap.Int("pts", remoteState.Pts),
		zap.Int("seq", remoteState.Seq),
		zap.Int("unread_count", remoteState.UnreadCount),
	)

	if err := m.syncCommon(ctx, remoteState.Pts); err != nil {
		return xerrors.Errorf("sync common: %w", err)
	}

	return nil
}

func (m *Manager) syncChannels(ctx context.Context) error {
	keys, err := m.storage.Keys(ctx)
	if err != nil {
		return xerrors.Errorf("get keys: %w", err)
	}

	for _, key := range keys {
		if key == commonStorageName {
			continue
		}

		var ch channelKey
		if err := ch.Parse(key); err != nil {
			m.log.Warn("Bad storage key", zap.Error(err))
			continue
		}

		input := tg.InputChannel(ch)
		if err := m.syncChannel(ctx, &input, 0); err != nil {
			m.log.Warn("Sync channel failed", zap.Error(err))
		}
	}

	return nil
}
