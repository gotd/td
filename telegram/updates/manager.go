package updates

import (
	"context"
	"sync"

	"github.com/go-faster/errors"
	"github.com/gotd/log"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
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
	state *internalState
	mux   sync.Mutex

	// immutable:

	cfg    Config
	lg     log.Helper
	tracer trace.Tracer
}

// New creates new manager.
func New(cfg Config) *Manager {
	cfg.setDefaults()
	return &Manager{
		cfg:    cfg,
		lg:     log.For(cfg.Logger),
		tracer: cfg.TracerProvider.Tracer(""),
	}
}

// Handle handles updates.
//
// Important:
// If Run method not called, all updates will be passed
// to the provided handler as-is without any order verification
// or short updates transformation.
func (m *Manager) Handle(ctx context.Context, u tg.UpdatesClass) error {
	ctx, span := m.tracer.Start(ctx, "updates.Manager.Handle")
	defer span.End()

	m.lg.Debug(ctx, "Handle")
	defer m.lg.Debug(ctx, "Handled")

	m.mux.Lock()
	state := m.state
	m.mux.Unlock()

	if state == nil {
		m.lg.Debug(ctx, "Handle (no internalState)")
		return m.cfg.Handler.Handle(ctx, u)
	}

	return state.Push(ctx, u)
}

type AuthOptions struct {
	IsBot   bool
	Forget  bool
	OnStart func(ctx context.Context)
}

// Run notifies manager about user authentication on the telegram server.
//
// If forget is true, local internalState (if exist) will be overwritten
// with remote internalState.
func (m *Manager) Run(ctx context.Context, api API, userID int64, opt AuthOptions) error {
	lg := m.lg.With(
		log.Int64("user_id", userID),
		log.Bool("is_bot", opt.IsBot),
		log.Bool("forget", opt.Forget),
	)
	lg.Debug(ctx, "Run")
	defer lg.Debug(ctx, "Done")

	wg, ctx := errgroup.WithContext(ctx)

	if err := func() error {
		m.mux.Lock()
		defer m.mux.Unlock()

		if m.state != nil {
			return errors.Errorf("already authorized (userID: %d)", m.state.selfID)
		}

		state, err := m.loadState(ctx, api, userID, opt.Forget)
		if err != nil {
			return errors.Wrap(err, "load internalState")
		}
		channels, err := m.loadChannels(ctx, userID)
		if err != nil {
			return errors.Wrap(err, "load channels")
		}

		diffLim := diffLimitUser
		if opt.IsBot {
			diffLim = diffLimitBot
		}

		m.state = newState(ctx, stateConfig{
			State:                 state,
			Channels:              channels,
			RawClient:             api,
			Tracer:                m.tracer,
			Logger:                m.cfg.Logger,
			Handler:               m.cfg.Handler,
			OnChannelTooLong:      m.cfg.OnChannelTooLong,
			OnChannelInaccessible: m.cfg.OnChannelInaccessible,
			OnTooLong:             m.cfg.OnTooLong,
			Storage:               m.cfg.Storage,
			Hasher:                m.cfg.AccessHasher,
			UserHasher:            m.cfg.UserAccessHasher,
			SelfID:                userID,
			DiffLimit:             diffLim,
			WorkGroup:             wg,
		})

		return nil
	}(); err != nil {
		return errors.Wrap(err, "setup")
	}
	if opt.OnStart != nil {
		opt.OnStart(ctx)
	}
	wg.Go(func() error {
		return m.state.Run(ctx)
	})
	lg.Debug(ctx, "Wait")
	return wg.Wait()
}

// loadChannels loads the locally stored state of the user's channels.
//
// Channels with a missing access hash are skipped and reported via
// Config.OnLoadChannelStateFailed.
func (m *Manager) loadChannels(ctx context.Context, userID int64) (map[int64]struct {
	Pts        int
	AccessHash int64
}, error) {
	channels := make(map[int64]struct {
		Pts        int
		AccessHash int64
	})
	if err := m.cfg.Storage.ForEachChannels(ctx, userID, func(ctx context.Context, channelID int64, pts int) error {
		hash, found, err := m.cfg.AccessHasher.GetChannelAccessHash(ctx, userID, channelID)
		if err != nil {
			return errors.Wrap(err, "get channel access hash")
		}

		if !found {
			m.cfg.OnLoadChannelStateFailed(channelID)
			return nil
		}

		channels[channelID] = struct {
			Pts        int
			AccessHash int64
		}{Pts: pts, AccessHash: hash}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "iterate channels")
	}
	return channels, nil
}

func (m *Manager) loadState(ctx context.Context, api API, userID int64, forget bool) (State, error) {
onNotFound:
	var state State
	if forget {
		remote, err := api.UpdatesGetState(ctx)
		if err != nil {
			return State{}, errors.Wrap(err, "get remote internalState")
		}

		state = state.fromRemote(remote)
		if err := m.cfg.Storage.SetState(ctx, userID, state); err != nil {
			return State{}, err
		}

		return state, nil
	}

	state, found, err := m.cfg.Storage.GetState(ctx, userID)
	if err != nil {
		return State{}, errors.Wrap(err, "restore local internalState")
	}

	if !found {
		m.cfg.OnLoadUserStateFailed()
		forget = true
		goto onNotFound
	}

	return state, nil
}

// Reset notifies manager about user logout.
func (m *Manager) Reset() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.state = nil
}
