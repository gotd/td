package updates

import (
	"context"
	"fmt"
	"github.com/gotd/td/clock"
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
	// trackQueue receives requests to track channels (Manager.TrackChannel).
	trackQueue chan trackRequest
	// done is closed when Run returns: nothing reads the queues anymore.
	done chan struct{}
	// runCtx is the context of channel workers.
	runCtx context.Context

	// Immutable fields.
	client                API
	log                   log.Helper
	handler               telegram.UpdateHandler
	stopOnApplyError      bool
	onTooLong             func(channelID int64)
	onChannelGap          func(gap ChannelGap)
	onChannelInaccessible func(channelID int64)
	onCommonTooLong       func()
	storage               StateStorage
	hasher                ChannelAccessHasher
	userHasher            UserAccessHasher
	selfID                int64
	diffLim               int
	wg                    *errgroup.Group
	tracer                trace.Tracer
	metrics               *updatesMetrics
	chDiffSem             chDiffSem
	clock                 clock.Clock
	// diffRetry retries a failed common difference (UPD-4).
	diffRetry *diffRetry
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
	Metrics               *updatesMetrics
	Handler               telegram.UpdateHandler
	StopOnApplyError      bool
	OnChannelTooLong      func(channelID int64)
	OnChannelGap          func(gap ChannelGap)
	OnChannelInaccessible func(channelID int64)
	OnTooLong             func()
	Storage               StateStorage
	Hasher                ChannelAccessHasher
	UserHasher            UserAccessHasher
	SelfID                int64
	DiffLimit             int
	WorkGroup             *errgroup.Group
	ChannelDiffSem        chDiffSem
	Clock                 clock.Clock
}

func newState(ctx context.Context, cfg stateConfig) *internalState {
	s := &internalState{
		externalQueue: make(chan tracedUpdate, 10),
		internalQueue: make(chan tracedUpdate, 10),
		affectedQueue: make(chan affectedUpdate, 10),
		trackQueue:    make(chan trackRequest),
		done:          make(chan struct{}),
		runCtx:        ctx,

		date:        cfg.State.Date,
		idleTimeout: time.NewTimer(idleTimeout),

		channels:      make(map[int64]*channelState),
		removeChannel: make(chan int64, 10),

		client:                cfg.RawClient,
		log:                   log.For(cfg.Logger),
		handler:               cfg.Handler,
		stopOnApplyError:      cfg.StopOnApplyError,
		onTooLong:             cfg.OnChannelTooLong,
		onChannelGap:          cfg.OnChannelGap,
		onChannelInaccessible: cfg.OnChannelInaccessible,
		onCommonTooLong:       cfg.OnTooLong,
		storage:               cfg.Storage,
		hasher:                cfg.Hasher,
		userHasher:            cfg.UserHasher,
		selfID:                cfg.SelfID,
		diffLim:               cfg.DiffLimit,
		wg:                    cfg.WorkGroup,
		tracer:                cfg.Tracer,
		metrics:               cfg.Metrics,
		chDiffSem:             cfg.ChannelDiffSem,
		clock:                 cfg.Clock,
		diffRetry:             newDiffRetry(cfg.Clock),
	}
	s.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Pts,
		Apply:        s.applyPts,
		Logger:       s.log.Named("pts").Logger(),
		Tracer:       s.tracer,
		Metrics:      s.metrics,
		Kind:         kindCommon,
		Name:         "pts",
	})
	s.qts = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Qts,
		Apply:        s.applyQts,
		Logger:       s.log.Named("qts").Logger(),
		Tracer:       s.tracer,
		Metrics:      s.metrics,
		Kind:         kindCommon,
		Name:         "qts",
	})
	s.seq = newSequenceBox(sequenceConfig{
		InitialState: cfg.State.Seq,
		Apply:        s.applySeq,
		Logger:       s.log.Named("seq").Logger(),
		Metrics:      s.metrics,
		Kind:         kindCommon,
		Name:         "seq",
	})

	for id, info := range cfg.Channels {
		s.startChannel(id, info.AccessHash, info.Pts)
		s.metrics.trackChannel(ctx, trackSourceStored, trackStarted)
	}
	if len(cfg.Channels) > 0 {
		s.log.Debug(ctx, "Tracking stored channels", log.Int("count", len(cfg.Channels)))
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
	case <-s.done:
		return ErrStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}

// stopped reports whether Run has returned.
func (s *internalState) stopped() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
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
	case <-s.done:
		return ErrStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}

// isFatalError returns true if error is fatal so we should stop updates handler.
func isFatalError(err error) bool {
	if isApplyError(err) {
		// Config.StopOnApplyError.
		return true
	}
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
	defer close(s.done)
	defer func() {
		// Channels of this state are no longer tracked.
		s.metrics.trackedDelta(context.Background(), -len(s.channels))
	}()
	if s.canceled(ctx) {
		return ctx.Err()
	}
	if err := s.getDifferenceLogger(ctx, "startup"); err != nil {
		return err
	}

	for {
		if s.canceled(ctx) {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			// With StopOnApplyError no new callbacks run after cancellation,
			// which may be caused by an apply error of a channel worker.
			if !s.stopOnApplyError && (len(s.pts.pending) > 0 || len(s.qts.pending) > 0 || len(s.seq.pending) > 0) {
				_ = s.getDifferenceLogger(ctx, "shutdown-flush-pending")
			}
			return ctx.Err()
		case u := <-s.externalQueue:
			if s.canceled(ctx) {
				return ctx.Err()
			}
			ctx := trace.ContextWithSpanContext(ctx, u.span)
			if err := s.handleUpdates(ctx, u.update); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
				if isFatalError(err) {
					return errors.Wrap(err, "fatal error")
				}
			}
		case u := <-s.internalQueue:
			if s.canceled(ctx) {
				return ctx.Err()
			}
			// Only channel workers write here, with other updates of
			// updates.getChannelDifference.
			ctx := withDifferenceSource(trace.ContextWithSpanContext(ctx, u.span))
			if err := s.handleUpdates(ctx, u.update); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
				if isFatalError(err) {
					return errors.Wrap(err, "fatal error")
				}
			}
		case a := <-s.affectedQueue:
			if s.canceled(ctx) {
				return ctx.Err()
			}
			ctx := trace.ContextWithSpanContext(ctx, a.span)
			if err := s.handleAffected(ctx, a.channelID, a.pts, a.ptsCount); err != nil {
				s.log.Error(ctx, "Handle affected error", log.Error(err))
				if isApplyError(err) {
					return err
				}
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug(ctx, "Pts gap timeout")
			if err := s.getDifferenceLogger(ctx, "pts-gap-timeout"); err != nil {
				return err
			}
		case <-s.qts.gapTimeout.C:
			s.log.Debug(ctx, "Qts gap timeout")
			if err := s.getDifferenceLogger(ctx, "qts-gap-timeout"); err != nil {
				return err
			}
		case <-s.seq.gapTimeout.C:
			s.log.Debug(ctx, "Seq gap timeout")
			if err := s.getDifferenceLogger(ctx, "seq-gap-timeout"); err != nil {
				return err
			}
		case <-s.idleTimeout.C:
			s.log.Debug(ctx, "Idle timeout")
			if err := s.getDifferenceLogger(ctx, "idle-timeout"); err != nil {
				return err
			}
		case <-s.diffRetry.C:
			s.diffRetry.fired()
			if err := s.getDifferenceLogger(ctx, "retry"); err != nil {
				return err
			}
		case channelID := <-s.removeChannel:
			s.removeChannelState(channelID)
		case r := <-s.trackQueue:
			r.result <- s.trackChannel(ctx, r, trackSourceAPI)
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
	s.metrics.trackedDelta(context.Background(), -1)
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
			return s.getDifference(ctx, "updates-peer-access-hash-unknown")
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
			return s.getDifference(ctx, "combined-peer-access-hash-unknown")
		}
		return s.handleSeq(ctx, u)
	case *tg.UpdateShort:
		return s.handleUpdates(ctx, &tg.UpdatesCombined{
			Updates: []tg.UpdateClass{u.Update},
			Date:    u.Date,
		})
	case *tg.UpdateShortMessage:
		if !s.shortMessagePeersKnown(ctx, u) {
			return s.getDifference(ctx, "short-message-peer-access-hash-unknown")
		}
		return s.handleUpdates(ctx, s.convertShortMessage(u))
	case *tg.UpdateShortChatMessage:
		if !s.shortChatMessagePeersKnown(ctx, u) {
			return s.getDifference(ctx, "short-chat-message-peer-access-hash-unknown")
		}
		return s.handleUpdates(ctx, s.convertShortChatMessage(u))
	case *tg.UpdateShortSentMessage:
		return s.handleUpdates(ctx, s.convertShortSentMessage(u))
	case *tg.UpdatesTooLong:
		return s.getDifference(ctx, "updates-too-long")
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
			return s.getDifference(ctx, "seq-zero-pts-changed")
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
		if err := s.dispatch(ctx, &tg.Updates{
			Updates: []tg.UpdateClass{u},
			Users:   ents.Users,
			Chats:   ents.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle updates error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
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
			if s.stopOnApplyError {
				// Do not start from the update: the stored checkpoint may be
				// behind it.
				return &applyError{errors.Wrap(err, "read channel checkpoint")}
			}
			localPts = pts - ptsCount
			s.log.Error(ctx, "GetChannelPts error", log.Error(err))
		}

		if !found {
			localPts = pts - ptsCount
			if err := s.storage.SetChannelPts(ctx, s.selfID, channelID, localPts); err != nil {
				s.log.Error(ctx, "SetChannelPts error", log.Error(err))
				if s.stopOnApplyError {
					return &applyError{err}
				}
			}
		}

		s.log.Debug(ctx, "Tracking channel on update",
			log.Int64("channel_id", channelID),
			log.Int("pts", localPts),
			log.Bool("stored_pts", found),
		)
		s.metrics.trackChannel(ctx, trackSourceUpdate, trackStarted)
		state = s.startChannel(channelID, accessHash, localPts)
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
		Hasher:                s.hasher,
		UserHasher:            s.userHasher,
		Handler:               s.handler,
		StopOnApplyError:      s.stopOnApplyError,
		OnChannelTooLong:      s.onTooLong,
		OnChannelGap:          s.onChannelGap,
		OnChannelInaccessible: s.onChannelInaccessible,
		RemoveChannel:         s.removeChannel,
		Logger:                s.log.Named("channel").With(log.Int64("channel_id", channelID)).Logger(),
		Tracer:                s.tracer,
		Metrics:               s.metrics,
		ChannelDiffSem:        s.chDiffSem,
		Clock:                 s.clock,
	})
}

func (s *internalState) getDifference(ctx context.Context, reason string) error {
	return s.getDifferenceChain(ctx, reason, false)
}

// getDifferenceChain fetches and applies the difference, continuing with
// further requests for slices and too-long differences.
//
// refetched is true once the chain made a follow-up request because a
// difference contained updatePtsChanged: the follow-up is made at most once per
// chain, so a server repeating updatePtsChanged cannot loop.
func (s *internalState) getDifferenceChain(ctx context.Context, reason string, refetched bool) error {
	if s.canceled(ctx) {
		return ctx.Err()
	}
	ctx, span := s.tracer.Start(ctx, "getDifference")
	defer span.End()

	s.resetIdleTimer()
	s.pts.gaps.Clear()
	s.qts.gaps.Clear()
	s.seq.gaps.Clear()

	s.log.Debug(ctx, "Getting difference",
		log.String("reason", reason),
		log.Int("pts", s.pts.State()),
		log.Int("qts", s.qts.State()),
		log.Int("seq", s.seq.State()),
		log.Int("date", s.date),
	)
	s.metrics.differenceTrigger(ctx, kindCommon, reason)
	// Updates of the difference are reported with "difference" source.
	ctx = withDifferenceSource(ctx)

	start := time.Now()
	diff, err := s.client.UpdatesGetDifference(ctx, &tg.UpdatesGetDifferenceRequest{
		Pts:  s.pts.State(),
		Qts:  s.qts.State(),
		Date: s.date,
	})
	if err != nil {
		s.metrics.difference(ctx, kindCommon, diffError, start)
		return errors.Wrap(err, "get difference")
	}
	result := commonDiffResult(diff)
	s.metrics.difference(ctx, kindCommon, result, start)
	if s.canceled(ctx) {
		// The response raced cancellation: do not apply it.
		return ctx.Err()
	}

	s.log.Debug(ctx, "Difference received",
		log.String("diff", fmt.Sprintf("%T", diff)),
		log.String("result", result),
		log.Duration("duration", time.Since(start)),
	)

	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		ptsChanged, err := s.applyDifference(ctx, differenceContents{
			NewMessages:          diff.NewMessages,
			NewEncryptedMessages: diff.NewEncryptedMessages,
			OtherUpdates:         diff.OtherUpdates,
			Users:                diff.Users,
			Chats:                diff.Chats,
			State:                diff.State,
		}, "updates.Difference")
		if err != nil {
			return err
		}
		if !ptsChanged {
			return nil
		}
		if refetched {
			s.log.Warn(ctx, "Ignoring repeated updatePtsChanged in difference",
				log.Int("pts", s.pts.State()),
			)
			return nil
		}
		s.log.Warn(ctx, "Got updatePtsChanged in difference, refetching",
			log.Int("pts", s.pts.State()),
		)
		return s.getDifferenceChain(ctx, "difference-pts-changed", true)

	// No events.
	case *tg.UpdatesDifferenceEmpty:
		if err := s.storage.SetDateSeq(ctx, s.selfID, diff.Date, diff.Seq); err != nil {
			s.log.Warn(ctx, "SetDateSeq error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
		}

		s.date = diff.Date
		s.seq.SetState(diff.Seq, "updates.differenceEmpty")
		return nil

	// Incomplete list of occurred events.
	case *tg.UpdatesDifferenceSlice:
		ptsChanged, err := s.applyDifference(ctx, differenceContents{
			NewMessages:          diff.NewMessages,
			NewEncryptedMessages: diff.NewEncryptedMessages,
			OtherUpdates:         diff.OtherUpdates,
			Users:                diff.Users,
			Chats:                diff.Chats,
			State:                diff.IntermediateState,
		}, "updates.differenceSlice")
		if err != nil {
			return err
		}
		if ptsChanged {
			// The continuation below is the refetch.
			s.log.Warn(ctx, "Got updatePtsChanged in difference slice",
				log.Int("pts", s.pts.State()),
				log.Bool("refetched", refetched),
			)
		}
		return s.getDifferenceChain(ctx, "difference-slice-continue", refetched || ptsChanged)

	// The difference is too long, and the specified internalState must be used to refetch updates.
	case *tg.UpdatesDifferenceTooLong:
		if err := s.storage.SetPts(ctx, s.selfID, diff.Pts); err != nil {
			s.log.Error(ctx, "SetPts error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
		}
		s.log.Warn(ctx, "Difference too long",
			log.Int("local_pts", s.pts.State()),
			log.Int("remote_pts", diff.Pts),
		)
		s.pts.SetState(diff.Pts, "updates.differenceTooLong")
		s.onCommonTooLong()
		return s.getDifferenceChain(ctx, "difference-too-long-continue", refetched)

	default:
		return errors.Errorf("unexpected diff type: %T", diff)
	}
}

// commonDiffResult returns the "result" metric attribute of diff.
func commonDiffResult(diff tg.UpdatesDifferenceClass) string {
	switch diff.(type) {
	case *tg.UpdatesDifferenceEmpty:
		return diffEmpty
	case *tg.UpdatesDifferenceSlice:
		return diffSlice
	case *tg.UpdatesDifferenceTooLong:
		return diffTooLong
	default:
		return diffFull
	}
}

// getDifferenceLogger calls getDifference, logging any error. It returns the
// error only if it must stop the manager (Config.StopOnApplyError).
//
// A failed difference is retried with backoff (diffRetry), and a state the
// server rejects with PERSISTENT_TIMESTAMP_INVALID or _EMPTY is reset from
// updates.getState, like TDLib does.
func (s *internalState) getDifferenceLogger(ctx context.Context, reason string) error {
	err := s.getDifference(ctx, reason)
	if err != nil && isPersistentTimestampReset(err) && ctx.Err() == nil {
		s.log.Warn(ctx, "Server rejected the local state, resetting from updates.getState",
			log.Error(err), log.String("reason", reason))
		err = s.resetState(ctx)
	}
	if err == nil {
		s.diffRetry.reset()
		return nil
	}
	if isApplyError(err) {
		s.log.Error(ctx, "get difference error", log.Error(err), log.String("reason", reason))
		return err
	}
	if !retryableDiffError(ctx, err) {
		s.log.Error(ctx, "get difference error", log.Error(err), log.String("reason", reason))
		return nil
	}
	delay := s.diffRetry.schedule(err)
	s.log.Error(ctx, "get difference error, retrying",
		log.Error(err),
		log.String("reason", reason),
		log.Duration("retry_in", delay),
	)
	return nil
}

// resetState replaces the common state with the one from updates.getState.
// The updates between the rejected state and the new one are lost, which is
// reported with OnTooLong. Stored channel states are kept: each field is
// stored on its own instead of with SetState.
func (s *internalState) resetState(ctx context.Context) error {
	st, err := s.client.UpdatesGetState(ctx)
	if err != nil {
		return errors.Wrap(err, "get state")
	}
	for _, set := range []func() error{
		func() error { return s.storage.SetPts(ctx, s.selfID, st.Pts) },
		func() error { return s.storage.SetQts(ctx, s.selfID, st.Qts) },
		func() error { return s.storage.SetDateSeq(ctx, s.selfID, st.Date, st.Seq) },
	} {
		if err := set(); err != nil {
			s.log.Error(ctx, "Store reset state error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
		}
	}
	s.pts.SetState(st.Pts, "updates.getState (reset)")
	s.qts.SetState(st.Qts, "updates.getState (reset)")
	s.seq.SetState(st.Seq, "updates.getState (reset)")
	s.date = st.Date
	s.onCommonTooLong()
	return nil
}

// canceled reports whether ctx is done and the manager must not run any new
// callbacks (Config.StopOnApplyError): the cancellation may be caused by an
// apply error of a sibling worker.
func (s *internalState) canceled(ctx context.Context) bool {
	return s.stopOnApplyError && ctx.Err() != nil
}

func (s *internalState) resetIdleTimer() {
	if len(s.idleTimeout.C) > 0 {
		<-s.idleTimeout.C
	}
	_ = s.idleTimeout.Reset(idleTimeout)
}
