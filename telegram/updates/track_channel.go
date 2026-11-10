package updates

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/log"
	"github.com/gotd/td/tg"
)

// trackRequest is a request to start tracking a channel.
type trackRequest struct {
	channelID  int64
	accessHash int64
	pts        int
	result     chan error
}

// TrackChannel starts tracking updates of the channel, if it is not tracked
// yet.
//
// The manager tracks only channels with a stored pts and access hash, and
// learns about other channels only from their next update, so messages posted
// before it are missed. Use TrackChannel to track channels joined while the
// client was offline, or channels reported via
// Config.OnLoadChannelStateFailed, e.g. with pts of the channel dialog from
// messages.getDialogs (Dialog.Pts).
//
// If the channel has a stored pts, tracking starts from it, so missed updates
// are fetched via updates.getChannelDifference. Otherwise it starts from the
// given pts, which must be positive.
//
// Can be called before Run or from AuthOptions.OnStart: such channels start
// being tracked when the manager starts, and errors are only logged.
func (m *Manager) TrackChannel(ctx context.Context, channelID, accessHash int64, pts int) error {
	r := trackRequest{
		channelID:  channelID,
		accessHash: accessHash,
		pts:        pts,
	}

	m.mux.Lock()
	state, started := m.state, m.started
	if state != nil && state.stopped() {
		// Run has stopped: the request would never be applied, as a new Run
		// requires Reset.
		m.mux.Unlock()
		return m.trackStopped(ctx)
	}
	if state == nil || !started {
		m.pendingTracks = append(m.pendingTracks, r)
		m.mux.Unlock()
		m.metrics.trackChannel(ctx, trackSourceAPI, trackPending)
		m.lg.Debug(ctx, "Channel will be tracked on Run", log.Int64("channel_id", channelID))
		return nil
	}
	m.mux.Unlock()

	r.result = make(chan error, 1)
	select {
	case state.trackQueue <- r:
	case <-state.done:
		return m.trackStopped(ctx)
	case <-ctx.Done():
		m.metrics.trackChannel(ctx, trackSourceAPI, trackCanceled)
		return ctx.Err()
	}
	select {
	case err := <-r.result:
		return err
	case <-state.done:
		// The request may have been applied right before Run stopped.
		select {
		case err := <-r.result:
			return err
		default:
			return m.trackStopped(ctx)
		}
	case <-ctx.Done():
		m.metrics.trackChannel(ctx, trackSourceAPI, trackCanceled)
		return ctx.Err()
	}
}

func (m *Manager) trackStopped(ctx context.Context) error {
	m.metrics.trackChannel(ctx, trackSourceAPI, trackStopped)
	m.metrics.stopped(ctx, opTrackChannel)
	return ErrStopped
}

// trackChannel starts tracking channel if it is not tracked yet.
//
// Must be called from the internalState goroutine (or before it starts).
func (s *internalState) trackChannel(ctx context.Context, r trackRequest, source string) error {
	if _, ok := s.channels[r.channelID]; ok {
		s.metrics.trackChannel(ctx, source, trackAlready)
		s.log.Debug(ctx, "Channel is already tracked", log.Int64("channel_id", r.channelID))
		return nil
	}
	if err := s.startTracking(ctx, r); err != nil {
		s.metrics.trackChannel(ctx, source, trackError)
		s.log.Warn(ctx, "Failed to track channel",
			log.Int64("channel_id", r.channelID),
			log.Error(err),
		)
		return err
	}
	s.metrics.trackChannel(ctx, source, trackStarted)
	return nil
}

func (s *internalState) startTracking(ctx context.Context, r trackRequest) error {
	localPts, found, err := s.storage.GetChannelPts(ctx, s.selfID, r.channelID)
	if err != nil {
		return errors.Wrap(err, "get channel pts")
	}
	if !found {
		if r.pts <= 0 {
			return errors.Errorf("channel %d: no stored pts and invalid pts %d", r.channelID, r.pts)
		}
		localPts = r.pts
		if err := s.storage.SetChannelPts(ctx, s.selfID, r.channelID, localPts); err != nil {
			return errors.Wrap(err, "set channel pts")
		}
	}
	if err := s.hasher.SetChannelAccessHash(ctx, s.selfID, r.channelID, r.accessHash); err != nil {
		return errors.Wrap(err, "set channel access hash")
	}

	s.log.Debug(ctx, "Tracking channel",
		log.Int64("channel_id", r.channelID),
		log.Int("pts", localPts),
		log.Bool("stored_pts", found),
	)
	s.startChannel(r.channelID, r.accessHash, localPts)
	return nil
}

// startChannel creates and runs channel state.
func (s *internalState) startChannel(channelID, accessHash int64, pts int) *channelState {
	state := s.newChannelState(channelID, accessHash, pts)
	s.channels[channelID] = state
	s.metrics.trackedDelta(s.runCtx, 1)
	ctx := s.runCtx
	s.wg.Go(func() error {
		return state.Run(ctx)
	})
	return state
}

// trackTooLong starts tracking a channel that is not tracked yet on
// updateChannelTooLong, if its stored pts and access hash are known.
//
// The error is non-nil only if it must stop the manager
// (Config.StopOnApplyError).
func (s *internalState) trackTooLong(ctx context.Context, u *tg.UpdateChannelTooLong, ents entities) (*channelState, bool, error) {
	localPts, found, err := s.storage.GetChannelPts(ctx, s.selfID, u.ChannelID)
	if err != nil {
		s.log.Error(ctx, "GetChannelPts error", log.Error(err))
		if s.stopOnApplyError {
			return nil, false, &applyError{errors.Wrap(err, "read channel checkpoint")}
		}
		return nil, false, nil
	}
	if !found {
		// No local state to recover from.
		return nil, false, nil
	}

	accessHash, found, err := s.hasher.GetChannelAccessHash(ctx, s.selfID, u.ChannelID)
	if err != nil {
		s.log.Error(ctx, "GetChannelAccessHash error", log.Error(err))
	}
	if !found {
		for _, c := range ents.Chats {
			ch, ok := c.(*tg.Channel)
			if !ok || ch.ID != u.ChannelID || ch.Min {
				continue
			}
			accessHash, found = ch.AccessHash, true
			if err := s.hasher.SetChannelAccessHash(ctx, s.selfID, u.ChannelID, accessHash); err != nil {
				s.log.Error(ctx, "SetChannelAccessHash error", log.Error(err))
			}
			break
		}
	}
	if !found {
		return nil, false, nil
	}

	s.log.Debug(ctx, "Tracking channel on ChannelTooLong",
		log.Int64("channel_id", u.ChannelID),
		log.Int("pts", localPts),
	)
	s.metrics.trackChannel(ctx, trackSourceTooLong, trackStarted)
	return s.startChannel(u.ChannelID, accessHash, localPts), true, nil
}
