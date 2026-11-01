package updates

import (
	"context"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/log"
)

type channelStateInfo struct {
	Pts        int
	AccessHash int64
}

type channelCompletion struct {
	state      *channelState
	deselected bool
}

var errChannelDeselected = errors.New("channel is not selected")

// RefreshChannelTracking schedules reconciliation with Config.TrackChannel.
// Publish the new selection before calling this method.
//
// It scans StateStorage.ForEachChannels and looks up hashes for selected saved
// channels in the calling goroutine. Call it outside storage transactions.
// A scan error leaves the pending reconciliation unchanged.
//
// A nil result means reconciliation was scheduled, not that channel workers
// have stopped or network catch-up has completed. Concurrent requests coalesce.
// Enabled saved channels catch up from stored PTS without needing a new push;
// missing access hashes are reported through OnLoadChannelStateFailed. Channels
// without saved PTS wait for normal update admission or caller bootstrap.
//
// Run must have initialized the manager; OnStart may call this method. With a
// nil TrackChannel it is a no-op. Cancellation after scheduling does not cancel
// the reconciliation or workers, which use Run's context.
func (m *Manager) RefreshChannelTracking(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mux.Lock()
	state := m.state
	if state == nil || state.runContext.Err() != nil {
		m.mux.Unlock()
		return errors.New("updates manager is not running")
	}
	if m.cfg.TrackChannel == nil {
		m.mux.Unlock()
		return nil
	}
	m.refreshRevision++
	revision := m.refreshRevision
	m.mux.Unlock()

	channels, err := m.loadChannels(ctx, state.selfID)
	if err != nil {
		return errors.Wrap(err, "refresh channel tracking")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	if m.state != state || state.runContext.Err() != nil {
		return errors.New("updates manager stopped during channel refresh")
	}
	if revision < m.refreshPublished {
		return nil
	}
	m.refreshPublished = revision
	// Only callers holding m.mux publish. The owner may consume at any time.
	select {
	case <-state.trackingQueue:
	default:
	}
	state.trackingQueue <- channels
	return nil
}

func (s *internalState) tracksChannel(id int64) bool {
	return s.trackChannel == nil || s.trackChannel(id)
}

// Only the common state owner changes registrations. The snapshot supplies
// candidates, not removals: live admission may have added a channel after scan.
func (s *internalState) reconcileChannels(ctx context.Context, channels map[int64]channelStateInfo) {
	for id, state := range s.channels {
		if !s.tracksChannel(id) {
			s.removeChannelState(state)
		}
	}
	for id, info := range channels {
		if !s.tracksChannel(id) || s.channels[id] != nil {
			continue
		}
		state := s.newChannelState(id, info.AccessHash, info.Pts)
		s.channels[id] = state
		s.startChannel(ctx, state, true)
	}
}

func (s *internalState) channelCompleted(ctx context.Context, completion channelCompletion) {
	state := completion.state
	if s.retiring[state.channelID] == state {
		delete(s.retiring, state.channelID)
	}
	if s.channels[state.channelID] != state {
		return
	}
	delete(s.channels, state.channelID)
	// A predicate change may stop admission before its explicit refresh reaches
	// the owner. If it was re-enabled meanwhile, do not leave a quiet channel
	// without a worker. Inaccessible channels are never restarted here.
	if completion.deselected && s.tracksChannel(state.channelID) {
		next := s.newChannelState(state.channelID, state.accessHash, state.pts.State())
		s.channels[state.channelID] = next
		s.startChannel(ctx, next, true)
	}
}

func (s *internalState) startChannel(ctx context.Context, state *channelState, restore bool) {
	if s.trackChannel == nil {
		s.wg.Go(func() error { return state.Run(ctx) })
		return
	}
	child, cancel := context.WithCancel(ctx)
	state.cancel = cancel
	state.persistenceContext = ctx
	previous := s.retiring[state.channelID]
	s.wg.Go(func() (retErr error) {
		started := false
		defer func() {
			deselected := state.deselected
			cancel()
			if !started {
				state.idleTimeout.Stop()
				state.pts.gapTimeout.Stop()
				close(state.done)
			}
			if ctx.Err() == nil && (errors.Is(retErr, errChannelDeselected) || retErr == context.Canceled) {
				retErr = nil
			}
			// Identity guards on the owner keep an old completion from removing
			// a replacement. Notify only after all effects and done have finished.
			select {
			case s.channelFinished <- channelCompletion{state: state, deselected: deselected}:
			case <-ctx.Done():
			}
		}()
		if previous != nil {
			// Even a canceled intermediate generation must carry this barrier:
			// a rapid disable/enable/disable/enable cannot overtake a late commit.
			<-previous.done
		}
		if err := state.admit(child); err != nil {
			return err
		}
		if restore || previous != nil {
			pts, found, err := s.storage.GetChannelPts(ctx, s.selfID, state.channelID)
			if err != nil || !found {
				s.log.Error(ctx, "Failed to restore selected channel checkpoint",
					log.Int64("channel_id", state.channelID), log.Error(err))
				s.onLoadChannelFailed(state.channelID)
				return nil
			}
			state.pts.SetState(pts, "channel-tracking-refresh")
		}
		started = true
		return state.Run(child)
	})
}

func (s *channelState) admit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.trackChannel != nil && !s.trackChannel(s.channelID) {
		s.deselected = true
		return errChannelDeselected
	}
	return nil
}

func (s *channelState) persistContext(ctx context.Context) context.Context {
	if s.persistenceContext == nil {
		return ctx
	}
	return trace.ContextWithSpan(s.persistenceContext, trace.SpanFromContext(ctx))
}
