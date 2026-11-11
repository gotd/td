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

// ErrStopped is returned by Handle, HandleAffected and TrackChannel after
// Run stopped, until Reset.
var ErrStopped = errors.New("updates manager stopped")

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
	// started is true when internalState goroutine is started.
	started bool
	// pendingTracks are TrackChannel requests made before start.
	pendingTracks []trackRequest
	// early are updates received before Run, with BeforeRunBuffer.
	early []tg.UpdatesClass
	// earlyDropped counts updates dropped before Run, for logging.
	earlyDropped int
	// replaying is true while Run replays early updates: updates received
	// meanwhile are buffered too, so they can't overtake buffered ones.
	replaying bool

	// immutable:

	cfg       Config
	lg        log.Helper
	tracer    trace.Tracer
	metrics   *updatesMetrics
	chDiffSem chDiffSem
}

// New creates new manager.
func New(cfg Config) *Manager {
	cfg.setDefaults()
	lg := log.For(cfg.Logger)
	return &Manager{
		cfg:       cfg,
		lg:        lg,
		tracer:    cfg.TracerProvider.Tracer(""),
		metrics:   newUpdatesMetrics(cfg.MeterProvider, lg),
		chDiffSem: newChDiffSem(cfg.MaxChannelDifferenceConcurrency),
	}
}

// Handle handles updates.
//
// Important:
// If Run method not called, updates are handled according to
// Config.BeforeRun: by default, they are passed to the provided handler as-is
// without any order verification or short updates transformation.
func (m *Manager) Handle(ctx context.Context, u tg.UpdatesClass) error {
	ctx, span := m.tracer.Start(ctx, "updates.Manager.Handle")
	defer span.End()

	m.lg.Debug(ctx, "Handle")
	defer m.lg.Debug(ctx, "Handled")

	m.mux.Lock()
	state := m.state
	if (state == nil && m.cfg.BeforeRun != BeforeRunPass) || m.replaying {
		// Under the lock, so Run takes every buffered update.
		action := earlyBuffered
		if m.cfg.BeforeRun == BeforeRunBuffer && len(m.early) < m.cfg.BeforeRunBufferSize {
			m.early = append(m.early, u)
		} else {
			m.earlyDropped++
			action = earlyDropped
		}
		replaying := m.replaying
		m.mux.Unlock()
		m.metrics.beforeRunAction(ctx, action, 1)
		m.lg.Debug(ctx, "Update received before Run",
			log.String("action", action),
			log.Bool("replaying", replaying),
		)
		return nil
	}
	m.mux.Unlock()

	if state == nil {
		m.lg.Debug(ctx, "Handle (no internalState)")
		m.metrics.beforeRunAction(ctx, earlyPassed, 1)
		return m.cfg.Handler.Handle(ctx, u)
	}

	if err := state.Push(ctx, u); err != nil {
		if errors.Is(err, ErrStopped) {
			m.metrics.stopped(ctx, opHandle)
		}
		return err
	}
	return nil
}

// HandleAffected applies the pts increment carried by a messages.affectedMessages
// or messages.affectedHistory RPC result, keeping the local pts in sync after a
// self-initiated read or delete (e.g. messages.readHistory, messages.deleteMessages).
//
// Without this, such results silently advance the server pts while the local pts
// stays behind, so the next genuine update looks like a gap and is postponed
// until the next pts-changing update or a getDifference (see issue #1382).
//
// channelID must be the channel the result belongs to, or 0 for the common
// (user/basic-group) sequence. The affected pts for a channel is applied only
// when that channel is already tracked. It is a no-op before Run is called.
//
// Most callers should not call this directly: use the hook middleware in
// telegram/updates/hook, which wires it from RPC results automatically.
func (m *Manager) HandleAffected(ctx context.Context, channelID int64, pts, ptsCount int) error {
	ctx, span := m.tracer.Start(ctx, "updates.Manager.HandleAffected")
	defer span.End()

	m.mux.Lock()
	state := m.state
	m.mux.Unlock()

	if state == nil {
		m.lg.Debug(ctx, "HandleAffected (no internalState)")
		return nil
	}

	if err := state.PushAffected(ctx, channelID, pts, ptsCount); err != nil {
		if errors.Is(err, ErrStopped) {
			m.metrics.stopped(ctx, opHandleAffected)
		}
		return err
	}
	return nil
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
	m.warnHashers(ctx, lg)

	wg, ctx := errgroup.WithContext(ctx)

	var (
		early       []tg.UpdatesClass
		dropped     int
		earlyRemote bool
	)
	if err := func() error {
		m.mux.Lock()
		defer m.mux.Unlock()

		if m.state != nil {
			return errors.Errorf("already authorized (userID: %d)", m.state.selfID)
		}

		state, remote, err := m.loadState(ctx, api, userID, opt.Forget)
		if err != nil {
			return errors.Wrap(err, "load internalState")
		}
		// Handle waits for m.mux during setup, so every buffered update was
		// received before the state was loaded.
		early, dropped = m.early, m.earlyDropped
		m.early, m.earlyDropped = nil, 0
		earlyRemote = remote
		m.replaying = len(early) > 0
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
			Metrics:               m.metrics,
			Logger:                m.cfg.Logger,
			Handler:               m.cfg.Handler,
			StopOnApplyError:      m.cfg.StopOnApplyError,
			OnChannelTooLong:      m.cfg.OnChannelTooLong,
			OnChannelGap:          m.cfg.OnChannelGap,
			OnChannelInaccessible: m.cfg.OnChannelInaccessible,
			OnTooLong:             m.cfg.OnTooLong,
			Storage:               m.cfg.Storage,
			Hasher:                m.cfg.AccessHasher,
			UserHasher:            m.cfg.UserAccessHasher,
			SelfID:                userID,
			DiffLimit:             diffLim,
			WorkGroup:             wg,
			ChannelDiffSem:        m.chDiffSem,
		})

		return nil
	}(); err != nil {
		return errors.Wrap(err, "setup")
	}
	if opt.OnStart != nil {
		opt.OnStart(ctx)
	}

	m.mux.Lock()
	state := m.state
	pending := m.pendingTracks
	m.pendingTracks = nil
	m.started = true
	m.mux.Unlock()
	if dropped > 0 {
		lg.Info(ctx, "Dropped updates received before Run", log.Int("count", dropped))
	}
	if len(early) > 0 {
		lg.Info(ctx, "Replaying updates received before Run",
			log.Int("count", len(early)),
			log.Bool("remote_state", earlyRemote),
		)
	}
	// internalState goroutine is not started yet, so it is safe to track
	// channels directly.
	for _, r := range pending {
		if err := state.trackChannel(ctx, r, trackSourceAPI); err != nil {
			lg.Error(ctx, "Track channel error",
				log.Int64("channel_id", r.channelID),
				log.Error(err),
			)
		}
	}

	wg.Go(func() error {
		return state.Run(ctx)
	})
	if len(early) > 0 {
		wg.Go(func() error {
			m.replayEarly(ctx, state, early, earlyRemote)
			return nil
		})
	}
	lg.Debug(ctx, "Wait")
	err := wg.Wait()

	m.mux.Lock()
	m.started = false
	m.mux.Unlock()
	return err
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

// loadState loads the user state and reports whether it was fetched from the
// server.
func (m *Manager) loadState(ctx context.Context, api API, userID int64, forget bool) (_ State, remote bool, _ error) {
onNotFound:
	var state State
	if forget {
		remote, err := api.UpdatesGetState(ctx)
		if err != nil {
			return State{}, false, errors.Wrap(err, "get remote internalState")
		}

		state = state.fromRemote(remote)
		if err := m.cfg.Storage.SetState(ctx, userID, state); err != nil {
			return State{}, false, err
		}

		return state, true, nil
	}

	state, found, err := m.cfg.Storage.GetState(ctx, userID)
	if err != nil {
		return State{}, false, errors.Wrap(err, "restore local internalState")
	}

	if !found {
		m.cfg.OnLoadUserStateFailed()
		forget = true
		goto onNotFound
	}

	return state, false, nil
}

// replayEarly handles updates buffered before Run (BeforeRunBuffer), then
// updates buffered while replaying, until the buffer is empty.
//
// If the state was fetched from the server, the updates buffered before Run
// are already covered by it: gap recovery would drop them as outdated and
// getDifference would not return them, so they are passed to Handler as-is.
func (m *Manager) replayEarly(ctx context.Context, state *internalState, early []tg.UpdatesClass, remote bool) {
	defer func() {
		m.mux.Lock()
		m.replaying = false
		m.mux.Unlock()
	}()
	for {
		for _, u := range early {
			if remote {
				m.metrics.beforeRunAction(ctx, earlyPassedRaw, 1)
				if err := m.cfg.Handler.Handle(ctx, u); err != nil {
					m.lg.Error(ctx, "Handle update received before Run", log.Error(err))
				}
				continue
			}
			// Push fails only when Run stops, and Run reports why.
			if err := state.Push(ctx, u); err != nil {
				m.lg.Debug(ctx, "Replay stopped", log.Error(err))
				return
			}
			m.metrics.beforeRunAction(ctx, earlyReplayed, 1)
		}
		// Updates received while replaying came after the state was loaded.
		remote = false

		m.mux.Lock()
		early = m.early
		m.early = nil
		if len(early) == 0 {
			m.replaying = false
			m.mux.Unlock()
			return
		}
		m.mux.Unlock()
	}
}

// Reset notifies manager about user logout.
func (m *Manager) Reset() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.state = nil
	m.started = false
	// Pending tracks and early updates belong to the previous user.
	m.pendingTracks = nil
	m.early = nil
	m.earlyDropped = 0
	m.replaying = false
}

// warnHashers warns when the update state is persistent but an access hasher
// is in memory: channel states are dropped on restart without their hashes,
// and users are forgotten.
func (m *Manager) warnHashers(ctx context.Context, lg log.Helper) {
	if _, mem := m.cfg.Storage.(*memStorage); mem {
		return
	}
	_, memChannels := m.cfg.AccessHasher.(*memAccessHasher)
	_, memUsers := m.cfg.UserAccessHasher.(*memUserAccessHasher)
	if memChannels || memUsers {
		lg.Warn(ctx, "Update state is persistent but access hashes are kept in memory: "+
			"set Config.AccessHasher (and UserAccessHasher) to a persistent hasher or *peers.Manager, "+
			"otherwise tracked channels are dropped on restart",
			log.Bool("memory_channel_hashes", memChannels),
			log.Bool("memory_user_hashes", memUsers),
		)
	}
}
