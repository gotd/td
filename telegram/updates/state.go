package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const (
	idleTimeout    = time.Minute * 15
	fastgapTimeout = time.Millisecond * 500

	diffLimitUser = 100
	diffLimitBot  = 100000
)

type tracedUpdate struct {
	update tg.UpdatesClass
	span   trace.SpanContext
}

// affectedUpdate carries a pts increment from a messages.affectedMessages /
// messages.affectedHistory RPC result to be applied on the internalState
// goroutine. A non-zero channelID routes it to that channel's pts sequence;
// zero targets the common sequence.
type affectedUpdate struct {
	channelID int64
	pts       int
	ptsCount  int
	span      trace.SpanContext
}

type internalState struct {
	// Updates channel.
	externalQueue chan tracedUpdate

	// Updates from channel states
	// during updates.getChannelDifference.
	internalQueue chan tracedUpdate

	// affectedQueue receives pts increments from messages.affected* RPC results
	// (see Manager.HandleAffected) to apply on the internalState goroutine.
	affectedQueue chan affectedUpdate

	// Common internalState.
	pts, qts, seq *sequenceBox
	date          int
	idleTimeout   *time.Timer

	// Channel states.
	channels map[int64]*channelState
	// removeChannel receives channel IDs from channel workers that lost access
	// (CHANNEL_PRIVATE) so they can be dropped from the channels map. Only the
	// internalState goroutine mutates the map, so this keeps access serialized.
	removeChannel chan int64

	// Immutable fields.
	client                API
	log                   log.Helper
	handler               telegram.UpdateHandler
	onTooLong             func(channelID int64)
	onChannelInaccessible func(channelID int64)
	onCommonTooLong       func()
	storage               StateStorage
	hasher                ChannelAccessHasher
	userHasher            UserAccessHasher
	selfID                int64
	diffLim               int
	wg                    *errgroup.Group
	tracer                trace.Tracer
}

type stateConfig struct {
	State    State
	Channels map[int64]struct {
		Pts        int
		AccessHash int64
	}
	RawClient             API
	Logger                log.Logger
	Tracer                trace.Tracer
	Handler               telegram.UpdateHandler
	OnChannelTooLong      func(channelID int64)
	OnChannelInaccessible func(channelID int64)
	OnTooLong             func()
	Storage               StateStorage
	Hasher                ChannelAccessHasher
	UserHasher            UserAccessHasher
	SelfID                int64
	DiffLimit             int
	WorkGroup             *errgroup.Group
}

func newState(ctx context.Context, cfg stateConfig) *internalState {
	s := &internalState{
		externalQueue: make(chan tracedUpdate, 10),
		internalQueue: make(chan tracedUpdate, 10),
		affectedQueue: make(chan affectedUpdate, 10),

		date:        cfg.State.Date,
		idleTimeout: time.NewTimer(idleTimeout),

		channels:      make(map[int64]*channelState),
		removeChannel: make(chan int64, 10),

		client:                cfg.RawClient,
		log:                   log.For(cfg.Logger),
		handler:               cfg.Handler,
		onTooLong:             cfg.OnChannelTooLong,
		onChannelInaccessible: cfg.OnChannelInaccessible,
		onCommonTooLong:       cfg.OnTooLong,
		storage:               cfg.Storage,
		hasher:                cfg.Hasher,
		userHasher:            cfg.UserHasher,
		selfID:                cfg.SelfID,
		diffLim:               cfg.DiffLimit,
		wg:                    cfg.WorkGroup,
		tracer:                cfg.Tracer,
	}
	s.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Pts,
		Apply:        s.applyPts,
		Logger:       s.log.Named("pts").Logger(),
		Tracer:       s.tracer,
	})
	s.qts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Qts,
		Apply:        s.applyQts,
		Logger:       s.log.Named("qts").Logger(),
		Tracer:       s.tracer,
	})
	s.seq = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Seq,
		Apply:        s.applySeq,
		Logger:       s.log.Named("seq").Logger(),
	})

	for id, info := range cfg.Channels {
		state := s.newChannelState(id, info.AccessHash, info.Pts)
		s.channels[id] = state
		s.wg.Go(func() error {
			return state.Run(ctx)
		})
	}

	return s
}

func (s *internalState) Push(ctx context.Context, u tg.UpdatesClass) error {
	tu := tracedUpdate{
		update: u,
		span:   trace.SpanContextFromContext(ctx),
	}
	select {
	case s.externalQueue <- tu:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PushAffected enqueues a pts increment from a messages.affected* RPC result.
func (s *internalState) PushAffected(ctx context.Context, channelID int64, pts, ptsCount int) error {
	au := affectedUpdate{
		channelID: channelID,
		pts:       pts,
		ptsCount:  ptsCount,
		span:      trace.SpanContextFromContext(ctx),
	}
	select {
	case s.affectedQueue <- au:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// isFatalError returns true if error is fatal so we should stop updates handler.
func isFatalError(err error) bool {
	// See https://github.com/gotd/td/issues/1458.
	if errors.Is(err, exchange.ErrKeyFingerprintNotFound) {
		return true
	}
	if tgerr.Is(err, "AUTH_KEY_UNREGISTERED", "SESSION_EXPIRED") {
		return true
	}
	if auth.IsUnauthorized(err) {
		return true
	}
	return false
}

func (s *internalState) Run(ctx context.Context) error {
	if s == nil {
		return errors.New("invalid: nil internalState")
	}
	s.log.Debug(ctx, "Starting updates handler")
	defer s.log.Debug(ctx, "Updates handler stopped")
	s.getDifferenceLogger(ctx)

	for {
		select {
		case <-ctx.Done():
			if len(s.pts.pending) > 0 || len(s.qts.pending) > 0 || len(s.seq.pending) > 0 {
				s.getDifferenceLogger(ctx)
			}
			return ctx.Err()
		case u := <-s.externalQueue:
			ctx := trace.ContextWithSpanContext(ctx, u.span)
			if err := s.handleUpdates(ctx, u.update); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
				if isFatalError(err) {
					return errors.Wrap(err, "fatal error")
				}
			}
		case u := <-s.internalQueue:
			ctx := trace.ContextWithSpanContext(ctx, u.span)
			if err := s.handleUpdates(ctx, u.update); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
				if isFatalError(err) {
					return errors.Wrap(err, "fatal error")
				}
			}
		case a := <-s.affectedQueue:
			ctx := trace.ContextWithSpanContext(ctx, a.span)
			if err := s.handleAffected(ctx, a.channelID, a.pts, a.ptsCount); err != nil {
				s.log.Error(ctx, "Handle affected error", log.Error(err))
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug(ctx, "Pts gap timeout")
			s.getDifferenceLogger(ctx)
		case <-s.qts.gapTimeout.C:
			s.log.Debug(ctx, "Qts gap timeout")
			s.getDifferenceLogger(ctx)
		case <-s.seq.gapTimeout.C:
			s.log.Debug(ctx, "Seq gap timeout")
			s.getDifferenceLogger(ctx)
		case <-s.idleTimeout.C:
			s.log.Debug(ctx, "Idle timeout")
			s.getDifferenceLogger(ctx)
		case channelID := <-s.removeChannel:
			s.removeChannelState(channelID)
		}
	}
}

// removeChannelState drops a channel from tracking after its worker stopped
// due to lost access (CHANNEL_PRIVATE). Must be called from the internalState
// goroutine, which is the sole mutator of the channels map.
func (s *internalState) removeChannelState(channelID int64) {
	if _, ok := s.channels[channelID]; !ok {
		return
	}
	delete(s.channels, channelID)
	s.log.Debug(context.Background(), "Removed inaccessible channel from tracking",
		log.Int64("channel_id", channelID),
	)
}

func (s *internalState) handleUpdates(ctx context.Context, u tg.UpdatesClass) error {
	ctx, span := s.tracer.Start(ctx, "handleUpdates")
	defer span.End()

	s.resetIdleTimer()

	switch u := u.(type) {
	case *tg.Updates:
		s.saveChannelHashes(ctx, u.Chats)
		s.saveUserHashes(ctx, u.Users)
		if !s.messageUpdatesPeersKnown(ctx, u.Updates) {
			return s.getDifference(ctx)
		}
		return s.handleSeq(ctx, &tg.UpdatesCombined{
			Updates:  u.Updates,
			Users:    u.Users,
			Chats:    u.Chats,
			Date:     u.Date,
			Seq:      u.Seq,
			SeqStart: u.Seq,
		})
	case *tg.UpdatesCombined:
		s.saveChannelHashes(ctx, u.Chats)
		s.saveUserHashes(ctx, u.Users)
		if !s.messageUpdatesPeersKnown(ctx, u.Updates) {
			return s.getDifference(ctx)
		}
		return s.handleSeq(ctx, u)
	case *tg.UpdateShort:
		return s.handleUpdates(ctx, &tg.UpdatesCombined{
			Updates: []tg.UpdateClass{u.Update},
			Date:    u.Date,
		})
	case *tg.UpdateShortMessage:
		if !s.shortMessagePeersKnown(ctx, u) {
			return s.getDifference(ctx)
		}
		return s.handleUpdates(ctx, s.convertShortMessage(u))
	case *tg.UpdateShortChatMessage:
		if !s.shortChatMessagePeersKnown(ctx, u) {
			return s.getDifference(ctx)
		}
		return s.handleUpdates(ctx, s.convertShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return s.handleUpdates(ctx, s.convertShortSentMessage(u))
	case *tg.UpdatesTooLong:
		return s.getDifference(ctx)
	default:
		panic(fmt.Sprintf("unexpected update type: %T", u))
	}
}

func (s *internalState) handleSeq(ctx context.Context, u *tg.UpdatesCombined) error {
	ctx, span := s.tracer.Start(ctx, "handleSeq")
	defer span.End()

	if err := validateSeq(u.Seq, u.SeqStart); err != nil {
		s.log.Error(ctx, "Seq validation failed", log.Error(err), log.Any("update", u))
		return nil
	}

	// Special case.
	if u.Seq == 0 {
		ptsChanged, err := s.applyCombined(ctx, u)
		if err != nil {
			return err
		}

		if ptsChanged {
			return s.getDifference(ctx)
		}

		return nil
	}

	return s.seq.Handle(ctx, update{
		Value: u,
		State: u.Seq,
		Count: u.Seq - u.SeqStart + 1,
	})
}

func (s *internalState) handlePts(ctx context.Context, pts, ptsCount int, u tg.UpdateClass, ents entities) error {
	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error(ctx, "Pts validation failed", log.Error(err), log.Any("update", u))
		return nil
	}

	return s.pts.Handle(ctx, update{
		Value:    u,
		State:    pts,
		Count:    ptsCount,
		Entities: ents,
	})
}

// handleAffected applies a pts increment from a messages.affectedMessages /
// messages.affectedHistory result. A non-zero channelID targets that channel's
// pts sequence (only when already tracked); zero targets the common sequence.
//
// A pts of 0 means the operation reported no pts and is ignored: feeding 0 to a
// sequenceBox would reset its state to 0 and desync everything.
func (s *internalState) handleAffected(ctx context.Context, channelID int64, pts, ptsCount int) error {
	if pts == 0 {
		return nil
	}

	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error(ctx, "Affected pts validation failed", log.Error(err),
			log.Int64("channel_id", channelID), log.Int("pts", pts), log.Int("pts_count", ptsCount))
		return nil
	}

	if channelID != 0 {
		st, ok := s.channels[channelID]
		if !ok {
			// Not tracked yet: skip. When the channel becomes tracked it runs its
			// own getChannelDifference, which reconciles pts from storage.
			s.log.Debug(ctx, "Affected pts for untracked channel, ignored",
				log.Int64("channel_id", channelID), log.Int("pts", pts))
			return nil
		}
		return st.Push(ctx, channelUpdate{
			affected: true,
			pts:      pts,
			ptsCount: ptsCount,
			span:     trace.SpanContextFromContext(ctx),
		})
	}

	return s.pts.Handle(ctx, update{
		Value: affectedPts{},
		State: pts,
		Count: ptsCount,
	})
}

func (s *internalState) handleQts(ctx context.Context, qts int, u tg.UpdateClass, ents entities) error {
	if err := validateQts(qts); err != nil {
		s.log.Error(ctx, "Qts validation failed", log.Error(err), log.Any("update", u))
		return nil
	}

	// A qts of 0 is not a sequence position: it is the "unset" sentinel carried
	// by, for example, bot business updates (updateBotNewBusinessMessage and
	// friends) delivered inside updates.difference. Feeding it to the qts
	// sequence resets the state to 0, which makes the next update look like a gap
	// from 0, triggers another getDifference returning the same qts=0 updates,
	// and loops forever re-dispatching them. Such updates carry no position to
	// order by, so dispatch them directly without touching the qts state.
	if qts == 0 {
		if err := s.handler.Handle(ctx, &tg.Updates{
			Updates: []tg.UpdateClass{u},
			Users:   ents.Users,
			Chats:   ents.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle updates error", log.Error(err))
		}

		return nil
	}

	return s.qts.Handle(ctx, update{
		Value:    u,
		State:    qts,
		Count:    1,
		Entities: ents,
	})
}

func (s *internalState) handleChannel(ctx context.Context, channelID int64, date, pts, ptsCount int, cu channelUpdate) error {
	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error(ctx, "Pts validation failed", log.Error(err), log.Any("update", cu.update))
		return nil
	}

	state, ok := s.channels[channelID]
	if !ok {
		accessHash, found, err := s.hasher.GetChannelAccessHash(context.Background(), s.selfID, channelID)
		if err != nil {
			s.log.Error(ctx, "GetChannelAccessHash error", log.Error(err))
		}

		if !found {
			if date == 0 {
				// Received update has no date field.
				date = s.date - 30
			} else {
				date-- // 1 sec back
			}

			// Try to get access hash from updates.getDifference.
			accessHash, found = s.restoreAccessHash(ctx, channelID, date)
			if !found {
				s.log.Debug(ctx, "Failed to recover missing access hash, update ignored",
					log.Int64("channel_id", channelID),
					log.Any("update", cu.update),
				)
				return nil
			}
		}

		localPts, found, err := s.storage.GetChannelPts(ctx, s.selfID, channelID)
		if err != nil {
			localPts = pts - ptsCount
			s.log.Error(ctx, "GetChannelPts error", log.Error(err))
		}

		if !found {
			localPts = pts - ptsCount
			if err := s.storage.SetChannelPts(ctx, s.selfID, channelID, localPts); err != nil {
				s.log.Error(ctx, "SetChannelPts error", log.Error(err))
			}
		}

		state = s.newChannelState(channelID, accessHash, localPts)
		s.channels[channelID] = state
		s.wg.Go(func() error {
			return state.Run(ctx)
		})
	}

	return state.Push(ctx, cu)
}

func (s *internalState) newChannelState(channelID, accessHash int64, initialPts int) *channelState {
	return newChannelState(channelStateConfig{
		Out:                   s.internalQueue,
		InitialPts:            initialPts,
		ChannelID:             channelID,
		AccessHash:            accessHash,
		SelfID:                s.selfID,
		Storage:               s.storage,
		DiffLimit:             s.diffLim,
		RawClient:             s.client,
		Handler:               s.handler,
		OnChannelTooLong:      s.onTooLong,
		OnChannelInaccessible: s.onChannelInaccessible,
		RemoveChannel:         s.removeChannel,
		Logger:                s.log.Named("channel").With(log.Int64("channel_id", channelID)).Logger(),
		Tracer:                s.tracer,
	})
}

func (s *internalState) getDifference(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "getDifference")
	defer span.End()

	s.resetIdleTimer()
	s.pts.gaps.Clear()
	s.qts.gaps.Clear()
	s.seq.gaps.Clear()

	s.log.Debug(ctx, "Getting difference")

	setState := func(state tg.UpdatesState, reason string) {
		if err := s.storage.SetState(ctx, s.selfID, State{}.fromRemote(&state)); err != nil {
			s.log.Warn(ctx, "SetState error", log.Error(err))
		}

		s.pts.SetState(state.Pts, reason)
		s.qts.SetState(state.Qts, reason)
		s.seq.SetState(state.Seq, reason)
		s.date = state.Date
	}

	diff, err := s.client.UpdatesGetDifference(ctx, &tg.UpdatesGetDifferenceRequest{
		Pts:  s.pts.State(),
		Qts:  s.qts.State(),
		Date: s.date,
	})
	if err != nil {
		return errors.Wrap(err, "get difference")
	}

	s.log.Debug(ctx, "Difference received", log.String("diff", fmt.Sprintf("%T", diff)))

	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		// Record recovered senders as known before dispatching, so a user
		// recovered via getDifference does not re-trigger getDifference for the
		// next short message from the same sender.
		s.saveUserHashes(ctx, diff.Users)

		if len(diff.OtherUpdates) > 0 {
			if err := s.handleUpdates(ctx, &tg.UpdatesCombined{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
			}); err != nil {
				return errors.Wrap(err, "handle diff.OtherUpdates")
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			if err := s.handler.Handle(ctx, &tg.Updates{
				Updates: append(
					msgsToUpdates(diff.NewMessages, false),
					encryptedMsgsToUpdates(diff.NewEncryptedMessages)...,
				),
				Users: diff.Users,
				Chats: diff.Chats,
			}); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
			}
		}

		setState(diff.State, "updates.Difference")
		return nil

	// No events.
	case *tg.UpdatesDifferenceEmpty:
		if err := s.storage.SetDateSeq(ctx, s.selfID, diff.Date, diff.Seq); err != nil {
			s.log.Warn(ctx, "SetDateSeq error", log.Error(err))
		}

		s.date = diff.Date
		s.seq.SetState(diff.Seq, "updates.differenceEmpty")
		return nil

	// Incomplete list of occurred events.
	case *tg.UpdatesDifferenceSlice:
		// Record recovered senders as known before dispatching, so a user
		// recovered via getDifference does not re-trigger getDifference for the
		// next short message from the same sender.
		s.saveUserHashes(ctx, diff.Users)

		if len(diff.OtherUpdates) > 0 {
			if err := s.handleUpdates(ctx, &tg.UpdatesCombined{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
				Date:    diff.IntermediateState.Date,
			}); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			if err := s.handler.Handle(ctx, &tg.Updates{
				Updates: append(
					msgsToUpdates(diff.NewMessages, false),
					encryptedMsgsToUpdates(diff.NewEncryptedMessages)...,
				),
				Users: diff.Users,
				Chats: diff.Chats,
			}); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
			}
		}

		setState(diff.IntermediateState, "updates.differenceSlice")
		return s.getDifference(ctx)

	// The difference is too long, and the specified internalState must be used to refetch updates.
	case *tg.UpdatesDifferenceTooLong:
		if err := s.storage.SetPts(ctx, s.selfID, diff.Pts); err != nil {
			s.log.Error(ctx, "SetPts error", log.Error(err))
		}
		s.pts.SetState(diff.Pts, "updates.differenceTooLong")
		s.onCommonTooLong()
		return s.getDifference(ctx)

	default:
		return errors.Errorf("unexpected diff type: %T", diff)
	}
}

func (s *internalState) getDifferenceLogger(ctx context.Context) {
	if err := s.getDifference(ctx); err != nil {
		s.log.Error(ctx, "get difference error", log.Error(err))
	}
}

func (s *internalState) resetIdleTimer() {
	if len(s.idleTimeout.C) > 0 {
		<-s.idleTimeout.C
	}
	_ = s.idleTimeout.Reset(idleTimeout)
}
