package updates

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/tg"
)

type channelUpdate struct {
	update tg.UpdateClass
	ctx    context.Context
	ents   entities
}

type channelState struct {
	// Updates from *state.
	uchan chan channelUpdate
	// Channel to pass diff.OtherUpdates into *state.
	outchan chan<- tg.UpdatesClass

	// Channel state.
	pts         *sequenceBox
	idleTimeout *time.Timer
	diffTimeout time.Time

	// Immutable fields.
	channelID  int64
	accessHash int64
	selfID     int64
	diffLim    int
	client     RawClient
	storage    StateStorage
	log        *zap.Logger
	handler    telegram.UpdateHandler
	onTooLong  func(channelID int64)

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

type channelStateConfig struct {
	Outchan          chan tg.UpdatesClass
	InitialPts       int
	ChannelID        int64
	AccessHash       int64
	SelfID           int64
	DiffLimit        int
	RawClient        RawClient
	Storage          StateStorage
	Handler          telegram.UpdateHandler
	OnChannelTooLong func(channelID int64)
	Logger           *zap.Logger
}

func newChannelState(cfg channelStateConfig) *channelState {
	ctx, cancel := context.WithCancel(context.Background())
	state := &channelState{
		uchan:   make(chan channelUpdate, 10),
		outchan: cfg.Outchan,

		idleTimeout: time.NewTimer(idleTimeout),

		channelID:  cfg.ChannelID,
		accessHash: cfg.AccessHash,
		selfID:     cfg.SelfID,
		diffLim:    cfg.DiffLimit,
		client:     cfg.RawClient,
		storage:    cfg.Storage,
		log:        cfg.Logger,
		handler:    cfg.Handler,
		onTooLong:  cfg.OnChannelTooLong,

		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	state.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.InitialPts,
		Apply:        state.applyPts,
		Logger:       cfg.Logger.Named("pts"),
	})

	return state
}

func (s *channelState) PushUpdate(u channelUpdate) { s.uchan <- u }

func (s *channelState) Run() {
	defer close(s.done)

	// Subscribe to channel updates.
	if err := s.getDifference(); err != nil {
		s.log.Error("Failed to subscribe to channel updates", zap.Error(err))
	}

	for {
		select {
		case u, ok := <-s.uchan:
			if !ok {
				if len(s.pts.pending) > 0 {
					s.getDifferenceLogerr()
				}
				return
			}

			if err := s.handleUpdate(u.ctx, u.update, u.ents); err != nil {
				s.log.Error("Handle update error", zap.Error(err))
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug("Gap timeout")
			s.getDifferenceLogerr()
		case <-s.idleTimeout.C:
			s.log.Debug("Idle timeout")
			s.resetIdleTimer()
			s.getDifferenceLogerr()
		}
	}
}

func (s *channelState) handleUpdate(ctx context.Context, u tg.UpdateClass, ents entities) error {
	s.resetIdleTimer()

	if long, ok := u.(*tg.UpdateChannelTooLong); ok {
		return s.handleTooLong(long)
	}

	channelID, pts, ptsCount, ok, err := isChannelPtsUpdate(u)
	if err != nil {
		return xerrors.Errorf("invalid update: %w", err)
	}

	if !ok {
		return xerrors.Errorf("expected channel update, got: %T", u)
	}

	if channelID != s.channelID {
		return xerrors.Errorf("update for wrong channel (channelID: %d)", channelID)
	}

	return s.pts.Handle(update{
		Value: u,
		State: pts,
		Count: ptsCount,
		Ents:  ents,
		Ctx:   ctx,
	})
}

func (s *channelState) handleTooLong(long *tg.UpdateChannelTooLong) error {
	remotePts, ok := long.GetPts()
	if !ok {
		s.log.Warn("Got UpdateChannelTooLong without pts field")
		return s.getDifference()
	}

	// Note: we still can fetch latest diffLim updates.
	// Should we do?
	if remotePts-s.pts.State() > s.diffLim {
		s.onTooLong(s.channelID)
		return nil
	}

	return s.getDifference()
}

func (s *channelState) applyPts(ctx context.Context, state int, updates []update) error {
	var (
		converted []tg.UpdateClass
		ents      entities
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Ents)
	}

	if err := s.handler.Handle(ctx, &tg.Updates{
		Updates: converted,
		Users:   ents.Users,
		Chats:   ents.Chats,
	}); err != nil {
		s.log.Error("Handle update error", zap.Error(err))
		return nil
	}

	if err := s.storage.SetChannelPts(s.selfID, s.channelID, state); err != nil {
		s.log.Error("SetChannelPts error", zap.Error(err))
	}

	return nil
}

func (s *channelState) getDifference() error {
	s.resetIdleTimer()
	s.pts.gaps.Clear()

	s.log.Debug("Getting difference")

	if now := time.Now(); now.Before(s.diffTimeout) {
		dur := s.diffTimeout.Sub(now)
		s.log.Debug("GetChannelDifference timeout", zap.Duration("duration", dur))
		if err := func() error {
			afterC := time.After(dur)
			for {
				select {
				case <-afterC:
					return nil
				case u, ok := <-s.uchan:
					if !ok {
						continue
					}

					// Ignoring updates to prevent *state worker from blocking.
					// All ignored updates should be restored by future getChannelDifference call.
					// At least I hope so...
					s.log.Debug("Ignoring update due to getChannelDifference timeout", zap.Any("update", u.update))
				case <-s.ctx.Done():
					return s.ctx.Err()
				}
			}
		}(); err != nil {
			return err
		}
	}

	diff, err := s.client.UpdatesGetChannelDifference(s.ctx, &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{
			ChannelID:  s.channelID,
			AccessHash: s.accessHash,
		},
		Filter: &tg.ChannelMessagesFilterEmpty{},
		Pts:    s.pts.State(),
		Limit:  s.diffLim,
	})
	if err != nil {
		return xerrors.Errorf("get channel difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesChannelDifference:
		if len(diff.OtherUpdates) > 0 {
			select {
			case s.outchan <- &tg.Updates{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
			}:
			case <-s.ctx.Done():
				return s.ctx.Err()
			}
		}

		if len(diff.NewMessages) > 0 {
			if err := s.handler.Handle(s.ctx, &tg.Updates{
				Updates: msgsToUpdates(diff.NewMessages),
				Users:   diff.Users,
				Chats:   diff.Chats,
			}); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		}

		if err := s.storage.SetChannelPts(s.selfID, s.channelID, diff.Pts); err != nil {
			s.log.Warn("SetChannelPts error", zap.Error(err))
		}

		s.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		if !diff.Final {
			return s.getDifference()
		}

		return nil

	case *tg.UpdatesChannelDifferenceEmpty:
		if err := s.storage.SetChannelPts(s.selfID, s.channelID, diff.Pts); err != nil {
			s.log.Warn("SetChannelPts error", zap.Error(err))
		}

		s.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		return nil

	case *tg.UpdatesChannelDifferenceTooLong:
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		remotePts, err := getDialogPts(diff.Dialog)
		if err != nil {
			s.log.Warn("UpdatesChannelDifferenceTooLong invalid Dialog", zap.Error(err))
		} else {
			if err := s.storage.SetChannelPts(s.selfID, s.channelID, remotePts); err != nil {
				s.log.Warn("SetChannelPts error", zap.Error(err))
			}

			s.pts.SetState(remotePts)
		}

		s.onTooLong(s.channelID)
		return nil

	default:
		return xerrors.Errorf("unexpected channel diff type: %T", diff)
	}
}

func (s *channelState) getDifferenceLogerr() {
	if err := s.getDifference(); err != nil {
		s.log.Error("get channel difference error", zap.Error(err))
	}
}

func (s *channelState) resetIdleTimer() {
	if len(s.idleTimeout.C) > 0 {
		<-s.idleTimeout.C
	}

	_ = s.idleTimeout.Reset(idleTimeout)
}

func (s *channelState) Close() {
	close(s.uchan)
	s.cancel()
	<-s.done
}
