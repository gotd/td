package updates

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/tg"
)

const (
	idleTimeout    = time.Minute * 15
	fastgapTimeout = time.Millisecond * 500

	diffLimitUser = 100
	diffLimitBot  = 100000
)

type updatesCtx struct {
	updates tg.UpdatesClass
	ctx     context.Context
}

type state struct {
	// Updates channel.
	externalQueue chan updatesCtx

	// Updates from channel states
	// during updates.getChannelDifference.
	internalQueue chan tg.UpdatesClass

	// Common state.
	pts, qts, seq *sequenceBox
	date          int
	idleTimeout   *time.Timer

	// Channel states.
	channels map[int64]*channelState

	// Immutable fields.
	client    RawClient
	log       *zap.Logger
	handler   telegram.UpdateHandler
	onTooLong func(channelID int64)
	storage   StateStorage
	hasher    ChannelAccessHasher
	selfID    int64
	diffLim   int

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

type stateConfig struct {
	State    State
	Channels map[int64]struct {
		Pts        int
		AccessHash int64
	}
	RawClient        RawClient
	Logger           *zap.Logger
	Handler          telegram.UpdateHandler
	OnChannelTooLong func(channelID int64)
	Storage          StateStorage
	Hasher           ChannelAccessHasher
	SelfID           int64
	DiffLimit        int
}

func newState(ctx context.Context, cfg stateConfig) *state {
	ctx, cancel := context.WithCancel(ctx)
	s := &state{
		externalQueue: make(chan updatesCtx, 10),
		internalQueue: make(chan tg.UpdatesClass, 10),

		date:        cfg.State.Date,
		idleTimeout: time.NewTimer(idleTimeout),

		channels: make(map[int64]*channelState),

		client:    cfg.RawClient,
		log:       cfg.Logger,
		handler:   cfg.Handler,
		onTooLong: cfg.OnChannelTooLong,
		storage:   cfg.Storage,
		hasher:    cfg.Hasher,
		selfID:    cfg.SelfID,
		diffLim:   cfg.DiffLimit,

		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	s.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Pts,
		Apply:        s.applyPts,
		Logger:       s.log.Named("pts"),
	})
	s.qts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Qts,
		Apply:        s.applyQts,
		Logger:       s.log.Named("qts"),
	})
	s.seq = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Seq,
		Apply:        s.applySeq,
		Logger:       s.log.Named("seq"),
	})

	for id, info := range cfg.Channels {
		state := s.newChannelState(id, info.AccessHash, info.Pts)
		s.channels[id] = state
		go state.Run()
	}

	return s
}

func (s *state) PushUpdates(ctx context.Context, u tg.UpdatesClass) {
	s.externalQueue <- updatesCtx{updates: u, ctx: ctx}
}

func (s *state) Run() {
	defer func() {
		for _, ch := range s.channels {
			ch.Close()
		}
		close(s.done)
	}()

	s.getDifferenceLogerr()

	for {
		select {
		case u, ok := <-s.externalQueue:
			if !ok {
				if len(s.pts.pending) > 0 || len(s.qts.pending) > 0 || len(s.seq.pending) > 0 {
					s.getDifferenceLogerr()
				}
				return
			}

			if err := s.handleUpdates(u.ctx, u.updates); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		case u := <-s.internalQueue:
			if err := s.handleUpdates(s.ctx, u); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug("Pts gap timeout")
			s.getDifferenceLogerr()
		case <-s.qts.gapTimeout.C:
			s.log.Debug("Qts gap timeout")
			s.getDifferenceLogerr()
		case <-s.seq.gapTimeout.C:
			s.log.Debug("Seq gap timeout")
			s.getDifferenceLogerr()
		case <-s.idleTimeout.C:
			s.log.Debug("Idle timeout")
			s.getDifferenceLogerr()
		}
	}
}

func (s *state) handleUpdates(ctx context.Context, u tg.UpdatesClass) error {
	s.resetIdleTimer()

	switch u := u.(type) {
	case *tg.Updates:
		s.saveChannelHashes(u.Chats)
		return s.handleSeq(ctx, &tg.UpdatesCombined{
			Updates:  u.Updates,
			Users:    u.Users,
			Chats:    u.Chats,
			Date:     u.Date,
			Seq:      u.Seq,
			SeqStart: u.Seq,
		})
	case *tg.UpdatesCombined:
		s.saveChannelHashes(u.Chats)
		return s.handleSeq(ctx, u)
	case *tg.UpdateShort:
		return s.handleUpdates(ctx, &tg.UpdatesCombined{
			Updates: []tg.UpdateClass{u.Update},
			Date:    u.Date,
		})
	case *tg.UpdateShortMessage:
		return s.handleUpdates(ctx, s.convertShortMessage(u))
	case *tg.UpdateShortChatMessage:
		return s.handleUpdates(ctx, s.convertShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return s.handleUpdates(ctx, s.convertShortSentMessage(u))
	case *tg.UpdatesTooLong:
		return s.getDifference()
	default:
		panic(fmt.Sprintf("unexpected update type: %T", u))
	}
}

func (s *state) handleSeq(ctx context.Context, u *tg.UpdatesCombined) error {
	if err := validateSeq(u.Seq, u.SeqStart); err != nil {
		s.log.Error("Seq validation failed", zap.Error(err), zap.Any("update", u))
		return nil
	}

	// Special case.
	if u.Seq == 0 {
		ptsChanged, err := s.applyCombined(ctx, u)
		if err != nil {
			return err
		}

		if ptsChanged {
			return s.getDifference()
		}

		return nil
	}

	return s.seq.Handle(update{
		Value: u,
		State: u.Seq,
		Count: u.Seq - u.SeqStart + 1,
	})
}

func (s *state) handlePts(pts, ptsCount int, u tg.UpdateClass, ents entities) error {
	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error("Pts validation failed", zap.Error(err), zap.Any("update", u))
		return nil
	}

	return s.pts.Handle(update{
		Value: u,
		State: pts,
		Count: ptsCount,
		Ents:  ents,
	})
}

func (s *state) handleQts(qts int, u tg.UpdateClass, ents entities) error {
	if err := validateQts(qts); err != nil {
		s.log.Error("Qts validation failed", zap.Error(err), zap.Any("update", u))
		return nil
	}

	return s.qts.Handle(update{
		Value: u,
		State: qts,
		Count: 1,
		Ents:  ents,
	})
}

func (s *state) handleChannel(channelID int64, date, pts, ptsCount int, cu channelUpdate) {
	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error("Pts validation failed", zap.Error(err), zap.Any("update", cu.update))
		return
	}

	state, ok := s.channels[channelID]
	if !ok {
		accessHash, found, err := s.hasher.GetChannelAccessHash(s.selfID, channelID)
		if err != nil {
			s.log.Error("GetChannelAccessHash error", zap.Error(err))
		}

		if !found {
			if date == 0 {
				// Received update has no date field.
				date = s.date - 30
			} else {
				date-- // 1 sec back
			}

			// Try to get access hash from updates.getDifference.
			accessHash, found = s.restoreAccessHash(channelID, date)
			if !found {
				s.log.Debug("Failed to recover missing access hash, update ignored",
					zap.Int64("channel_id", channelID),
					zap.Any("update", cu.update),
				)
				return
			}
		}

		localPts, found, err := s.storage.GetChannelPts(s.selfID, channelID)
		if err != nil {
			localPts = pts - ptsCount
			s.log.Error("GetChannelPts error", zap.Error(err))
		}

		if !found {
			localPts = pts - ptsCount
			if err := s.storage.SetChannelPts(s.selfID, channelID, localPts); err != nil {
				s.log.Error("SetChannelPts error", zap.Error(err))
			}
		}

		state = s.newChannelState(channelID, accessHash, localPts)
		s.channels[channelID] = state
		go state.Run()
	}

	state.PushUpdate(cu)
}

func (s *state) newChannelState(channelID, accessHash int64, initialPts int) *channelState {
	return newChannelState(channelStateConfig{
		Outchan:          s.internalQueue,
		InitialPts:       initialPts,
		ChannelID:        channelID,
		AccessHash:       accessHash,
		SelfID:           s.selfID,
		Storage:          s.storage,
		DiffLimit:        s.diffLim,
		RawClient:        s.client,
		Handler:          s.handler,
		OnChannelTooLong: s.onTooLong,
		Logger:           s.log.Named("channel").With(zap.Int64("channel_id", channelID)),
	})
}

func (s *state) getDifference() error {
	s.resetIdleTimer()
	s.pts.gaps.Clear()
	s.qts.gaps.Clear()
	s.seq.gaps.Clear()

	s.log.Debug("Getting difference")

	setState := func(state tg.UpdatesState) {
		if err := s.storage.SetState(s.selfID, State{}.fromRemote(&state)); err != nil {
			s.log.Warn("SetState error", zap.Error(err))
		}

		s.pts.SetState(state.Pts)
		s.qts.SetState(state.Qts)
		s.seq.SetState(state.Seq)
		s.date = state.Date
	}

	diff, err := s.client.UpdatesGetDifference(s.ctx, &tg.UpdatesGetDifferenceRequest{
		Pts:  s.pts.State(),
		Qts:  s.qts.State(),
		Date: s.date,
	})
	if err != nil {
		return xerrors.Errorf("get difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		if len(diff.OtherUpdates) > 0 {
			if err := s.handleUpdates(s.ctx, &tg.UpdatesCombined{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
			}); err != nil {
				return xerrors.Errorf("handle diff.OtherUpdates: %w", err)
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			if err := s.handler.Handle(s.ctx, &tg.Updates{
				Updates: append(
					msgsToUpdates(diff.NewMessages),
					encryptedMsgsToUpdates(diff.NewEncryptedMessages)...,
				),
				Users: diff.Users,
				Chats: diff.Chats,
			}); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		}

		setState(diff.State)
		return nil

	// No events.
	case *tg.UpdatesDifferenceEmpty:
		if err := s.storage.SetDateSeq(s.selfID, diff.Date, diff.Seq); err != nil {
			s.log.Warn("SetDateSeq error", zap.Error(err))
		}

		s.date = diff.Date
		s.seq.SetState(diff.Seq)
		return nil

	// Incomplete list of occurred events.
	case *tg.UpdatesDifferenceSlice:
		if len(diff.OtherUpdates) > 0 {
			if err := s.handleUpdates(s.ctx, &tg.UpdatesCombined{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
				Date:    diff.IntermediateState.Date,
			}); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			if err := s.handler.Handle(s.ctx, &tg.Updates{
				Updates: append(
					msgsToUpdates(diff.NewMessages),
					encryptedMsgsToUpdates(diff.NewEncryptedMessages)...,
				),
				Users: diff.Users,
				Chats: diff.Chats,
			}); err != nil {
				s.log.Error("Handle updates error", zap.Error(err))
			}
		}

		setState(diff.IntermediateState)
		return s.getDifference()

	// The difference is too long, and the specified state must be used to refetch updates.
	case *tg.UpdatesDifferenceTooLong:
		if err := s.storage.SetPts(s.selfID, diff.Pts); err != nil {
			s.log.Error("SetPts error", zap.Error(err))
		}
		s.pts.SetState(diff.Pts)
		return s.getDifference()

	default:
		return xerrors.Errorf("unexpected diff type: %T", diff)
	}
}

func (s *state) getDifferenceLogerr() {
	if err := s.getDifference(); err != nil {
		s.log.Error("get difference error", zap.Error(err))
	}
}

func (s *state) resetIdleTimer() {
	if len(s.idleTimeout.C) > 0 {
		<-s.idleTimeout.C
	}
	_ = s.idleTimeout.Reset(idleTimeout)
}

func (s *state) Close() {
	close(s.externalQueue)
	s.cancel()
	<-s.done
}
