package updates

import (
	"context"
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/tg"
)

var _ telegram.UpdateHandler = (*Manager)(nil)

// Manager deals with gaps.
//
// Important:
// Updates produced by this manager may contain
// negative Pts/Qts/Seq values in tg.UpdateClass/tg.UpdatesClass
// (does not affects to the tg.MessageClass).
//
// This is because telegram server does not return these sequences
// for getDifference/getChannelDifference results.
// You SHOULD NOT use them in update handlers at all.
type Manager struct {
	state *state
	mux   sync.Mutex

	cfg Config // immutable
}

// New creates new manager.
func New(cfg Config) *Manager {
	cfg.setDefaults()
	return &Manager{cfg: cfg}
}

// Handle handles updates.
//
// Important:
// If Auth method not called, all updates will be passed
// to the provided handler as-is without any order verification
// or short updates transformation.
func (m *Manager) Handle(ctx context.Context, u tg.UpdatesClass) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state == nil {
		return m.cfg.Handler.Handle(ctx, u)
	}

	m.state.PushUpdates(ctx, u)
	return nil
}

// Auth notifies manager about user authentication on the telegram server.
//
// If forget is true, local state (if exist) will be overwritten
// with remote state.
func (m *Manager) Auth(ctx context.Context, client RawClient, userID int64, isBot, forget bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state != nil {
		return xerrors.Errorf("already authorized (userID: %d)", m.state.selfID)
	}

	state, err := m.loadState(client, userID, forget)
	if err != nil {
		return xerrors.Errorf("load state: %w", err)
	}

	channels := make(map[int64]struct {
		Pts        int
		AccessHash int64
	})
	if err := m.cfg.Storage.ForEachChannels(userID, func(channelID int64, pts int) error {
		hash, found, err := m.cfg.AccessHasher.GetChannelAccessHash(userID, channelID)
		if err != nil {
			return err
		}

		if !found {
			return nil
		}

		channels[channelID] = struct {
			Pts        int
			AccessHash int64
		}{pts, hash}
		return nil
	}); err != nil {
		return err
	}

	diffLim := diffLimitUser
	if isBot {
		diffLim = diffLimitBot
	}

	m.state = newState(ctx, stateConfig{
		State:            state,
		Channels:         channels,
		RawClient:        client,
		Logger:           m.cfg.Logger,
		Handler:          m.cfg.Handler,
		OnChannelTooLong: m.cfg.OnChannelTooLong,
		Storage:          m.cfg.Storage,
		Hasher:           m.cfg.AccessHasher,
		SelfID:           userID,
		DiffLimit:        diffLim,
	})
	go m.state.Run()
	return nil
}

func (m *Manager) loadState(client RawClient, userID int64, forget bool) (State, error) {
onNotFound:
	var state State
	if forget {
		remote, err := client.UpdatesGetState(context.TODO())
		if err != nil {
			return State{}, xerrors.Errorf("get remote state: %w", err)
		}

		state = state.fromRemote(remote)
		if err := m.cfg.Storage.SetState(userID, state); err != nil {
			return State{}, err
		}

		if err := m.cfg.Storage.SetState(userID, state); err != nil {
			return State{}, err
		}

		return state, nil
	}

	state, found, err := m.cfg.Storage.GetState(userID)
	if err != nil {
		return State{}, xerrors.Errorf("restore local state: %w", err)
	}

	if !found {
		forget = true
		goto onNotFound
	}

	return state, nil
}

// Logout notifies manager about user logout.
func (m *Manager) Logout() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state == nil {
		return xerrors.New("not authorized, nothing to do")
	}

	m.state.Close()
	m.state = nil
	return nil
}
