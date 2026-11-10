package updates

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// errChannelInaccessible is a sentinel returned by getDifference when the
// account has lost access to the channel (CHANNEL_PRIVATE). It signals the
// Run loop to stop the worker instead of logging the error and retrying.
var errChannelInaccessible = errors.New("channel is inaccessible")

type channelUpdate struct {
	update   tg.UpdateClass
	entities entities
	span     trace.SpanContext

	// affected, when true, makes the worker advance its pts by (pts, ptsCount)
	// without dispatching, applying a messages.affected* result for this channel
	// (see Manager.HandleAffected). update is nil in this case.
	affected bool
	pts      int
	ptsCount int
}

type channelState struct {
	// Updates from *internalState.
	updates chan channelUpdate
	// Channel to pass diff.OtherUpdates into *internalState.
	out chan<- tracedUpdate

	// Channel internalState.
	pts         *sequenceBox
	idleTimeout *time.Timer
	diffTimeout time.Time

	// done is closed when Run returns, so that Push never blocks on a stopped
	// worker (e.g. after the channel became inaccessible and is pending removal).
	done chan struct{}

	// Immutable fields.
	channelID        int64
	accessHash       int64
	selfID           int64
	diffLim          int
	client           API
	storage          StateStorage
	hasher           ChannelAccessHasher
	userHasher       UserAccessHasher
	log              log.Helper
	tracer           trace.Tracer
	metrics          *updatesMetrics
	handler          telegram.UpdateHandler
	stopOnApplyError bool
	onTooLong        func(channelID int64)
	onGap            func(gap ChannelGap)
	onInaccessible   func(channelID int64)
	// removeChannel signals *internalState to drop this channel from tracking.
	removeChannel chan<- int64
	// chDiffSem bounds concurrent getChannelDifference calls across all channels
	// of the manager. nil means unlimited.
	chDiffSem chDiffSem
}

// ChannelGap describes a channel gap that the manager cannot recover
// (updates.channelDifferenceTooLong).
type ChannelGap struct {
	// ChannelID is the channel ID.
	ChannelID int64
	// AccessHash is the channel access hash.
	AccessHash int64
	// LocalPts is the local pts before the gap.
	LocalPts int
	// RemotePts is the new pts after the gap (pts of the channel dialog).
	RemotePts int
	// TopMessage is the ID of the latest channel message (from the
	// channel dialog), zero if unknown.
	TopMessage int
	// Messages are the latest channel messages returned with the gap. They
	// may include messages delivered to Handler before the gap, so they are
	// not passed to Handler.
	Messages []tg.MessageClass
	// Users and Chats are entities of Messages.
	Users []tg.UserClass
	Chats []tg.ChatClass
}

type channelStateConfig struct {
	Out                   chan tracedUpdate
	InitialPts            int
	ChannelID             int64
	AccessHash            int64
	SelfID                int64
	DiffLimit             int
	RawClient             API
	Storage               StateStorage
	Hasher                ChannelAccessHasher
	UserHasher            UserAccessHasher
	Handler               telegram.UpdateHandler
	StopOnApplyError      bool
	OnChannelTooLong      func(channelID int64)
	OnChannelGap          func(gap ChannelGap)
	OnChannelInaccessible func(channelID int64)
	RemoveChannel         chan<- int64
	Logger                log.Logger
	Tracer                trace.Tracer
	Metrics               *updatesMetrics
	ChannelDiffSem        chDiffSem
}

func newChannelState(cfg channelStateConfig) *channelState {
	state := &channelState{
		updates: make(chan channelUpdate, 10),
		out:     cfg.Out,

		idleTimeout: time.NewTimer(idleTimeout),
		done:        make(chan struct{}),

		channelID:        cfg.ChannelID,
		accessHash:       cfg.AccessHash,
		selfID:           cfg.SelfID,
		diffLim:          cfg.DiffLimit,
		client:           cfg.RawClient,
		storage:          cfg.Storage,
		hasher:           cfg.Hasher,
		userHasher:       cfg.UserHasher,
		log:              log.For(cfg.Logger),
		handler:          cfg.Handler,
		stopOnApplyError: cfg.StopOnApplyError,
		onTooLong:        cfg.OnChannelTooLong,
		onGap:            cfg.OnChannelGap,
		onInaccessible:   cfg.OnChannelInaccessible,
		removeChannel:    cfg.RemoveChannel,
		tracer:           cfg.Tracer,
		metrics:          cfg.Metrics,
		chDiffSem:        cfg.ChannelDiffSem,
	}

	state.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.InitialPts,
		Apply:        state.applyPts,
		Logger:       log.Named(cfg.Logger, "pts"),
		Tracer:       cfg.Tracer,
		Metrics:      cfg.Metrics,
		Kind:         kindChannel,
		Name:         "pts",
	})

	return state
}

func (s *channelState) Push(ctx context.Context, u channelUpdate) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		// Worker has stopped (channel inaccessible, pending removal): drop the
		// update instead of blocking. It is restored by a future subscribe if
		// access is regained.
		return nil
	case s.updates <- u:
		return nil
	}
}

func (s *channelState) Run(ctx context.Context) error {
	defer close(s.done)

	// Subscribe to channel updates.
	if s.canceled(ctx) {
		return ctx.Err()
	}
	if err := s.getDifference(ctx, "channel-subscribe"); err != nil {
		if errors.Is(err, errChannelInaccessible) {
			return nil
		}
		s.log.Error(ctx, "Failed to subscribe to channel updates", log.Error(err))
		if isApplyError(err) {
			return err
		}
	}

	for {
		if s.canceled(ctx) {
			return ctx.Err()
		}
		select {
		case u := <-s.updates:
			if s.canceled(ctx) {
				return ctx.Err()
			}
			ctx := trace.ContextWithSpanContext(ctx, u.span)
			handle := s.handleUpdate
			if u.affected {
				handle = func(ctx context.Context, _ tg.UpdateClass, _ entities) error {
					return s.handleAffected(ctx, u.pts, u.ptsCount)
				}
			}
			if err := handle(ctx, u.update, u.entities); err != nil {
				if errors.Is(err, errChannelInaccessible) {
					return nil
				}
				s.log.Error(ctx, "Handle update error", log.Error(err))
				if isApplyError(err) {
					return err
				}
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug(ctx, "Gap timeout")
			if stop, err := s.getDifferenceLogger(ctx, "channel-pts-gap-timeout"); stop {
				return err
			}
		case <-ctx.Done():
			// With StopOnApplyError no new callbacks run after cancellation,
			// which may be caused by an apply error of another worker.
			if !s.stopOnApplyError && len(s.pts.pending) > 0 {
				// This will probably fail.
				_, _ = s.getDifferenceLogger(ctx, "channel-shutdown-flush-pending")
			}
			return ctx.Err()
		case <-s.idleTimeout.C:
			// The server timeout paces idle polling, not gap recovery or
			// non-final differences. Keep handling updates while it elapses.
			if remaining := time.Until(s.diffTimeout); remaining > 0 {
				s.idleTimeout.Reset(remaining)
				continue
			}
			s.log.Debug(ctx, "Idle timeout")
			s.resetIdleTimer()
			if stop, err := s.getDifferenceLogger(ctx, "channel-idle-timeout"); stop {
				return err
			}
		}
	}
}

func (s *channelState) handleUpdate(ctx context.Context, u tg.UpdateClass, ents entities) error {
	ctx, span := s.tracer.Start(ctx, "channelState.handleUpdate")
	defer span.End()

	s.resetIdleTimer()

	if long, ok := u.(*tg.UpdateChannelTooLong); ok {
		return s.handleTooLong(ctx, long)
	}

	channelID, pts, ptsCount, ok, err := tg.IsChannelPtsUpdate(u)
	if err != nil {
		return errors.Wrap(err, "invalid update")
	}

	if !ok {
		return errors.Errorf("expected channel update, got: %T", u)
	}

	if channelID != s.channelID {
		return errors.Errorf("update for wrong channel (channelID: %d)", channelID)
	}

	return s.pts.Handle(ctx, update{
		Value:    u,
		State:    pts,
		Count:    ptsCount,
		Entities: ents,
	})
}

// handleAffected applies a pts increment from a messages.affected* result to
// this channel's pts sequence without dispatching. A pts of 0 is ignored: it
// would reset the sequence state to 0 (see internalState.handleAffected).
func (s *channelState) handleAffected(ctx context.Context, pts, ptsCount int) error {
	if pts == 0 {
		return nil
	}
	if err := validatePts(pts, ptsCount); err != nil {
		s.log.Error(ctx, "Affected pts validation failed", log.Error(err),
			log.Int("pts", pts), log.Int("pts_count", ptsCount))
		return nil
	}
	return s.pts.Handle(ctx, update{
		Value: affectedPts{},
		State: pts,
		Count: ptsCount,
	})
}

func (s *channelState) handleTooLong(ctx context.Context, long *tg.UpdateChannelTooLong) error {
	ctx, span := s.tracer.Start(ctx, "channelState.handleTooLong")
	defer span.End()

	if _, ok := long.GetPts(); !ok {
		s.log.Warn(ctx, "Got UpdateChannelTooLong without pts field")
		return s.getDifference(ctx, "channel-too-long-no-pts")
	}

	// Even if the gap is larger than diffLim, fetch the difference: the server
	// returns channelDifferenceTooLong with the latest messages, which are
	// dispatched, and the gap is reported via OnChannelTooLong/OnChannelGap.
	return s.getDifference(ctx, "channel-too-long")
}

func (s *channelState) applyPts(ctx context.Context, state int, updates []update) error {
	ctx, span := s.tracer.Start(ctx, "channelState.applyPts")
	defer span.End()

	var (
		converted []tg.UpdateClass
		ents      entities
	)

	for _, update := range updates {
		// affectedPts is a pts-only marker (see Manager.HandleAffected): it
		// advances the sequence but has nothing to dispatch.
		if _, ok := update.Value.(affectedPts); ok {
			continue
		}
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Entities)
	}

	if len(converted) > 0 {
		if err := s.dispatch(ctx, &tg.Updates{
			Updates: converted,
			Users:   ents.Users,
			Chats:   ents.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle update error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
			return nil
		}
	}

	if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, state); err != nil {
		s.log.Error(ctx, "SetChannelPts error", log.Error(err))
		if s.stopOnApplyError {
			return &applyError{err}
		}
	}

	return nil
}

func (s *channelState) getDifference(ctx context.Context, reason string) error {
	if s.canceled(ctx) {
		return ctx.Err()
	}
	ctx, span := s.tracer.Start(ctx, "channelState.getDifference")
	defer span.End()
	s.pts.gaps.Clear()

	s.log.Debug(ctx, "Getting difference",
		log.String("reason", reason),
		log.Int64("channel_id", s.channelID),
		log.Int("pts", s.pts.State()),
	)
	s.metrics.differenceTrigger(ctx, kindChannel, reason)
	// Updates of the difference are reported with "difference" source.
	ctx = withDifferenceSource(ctx)

	// Bound concurrent getChannelDifference across channels: a member of many
	// active channels would otherwise burst past the per-account method limit.
	// The slot wraps only the RPC — released before processing and the !Final
	// recursion so a busy channel's catch-up does not monopolise it (and so a
	// limit of 1 cannot self-deadlock on the nested call).
	if err := s.chDiffSem.acquire(ctx); err != nil {
		return err
	}
	start := time.Now()
	diff, err := s.client.UpdatesGetChannelDifference(ctx, &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{
			ChannelID:  s.channelID,
			AccessHash: s.accessHash,
		},
		Filter: &tg.ChannelMessagesFilterEmpty{},
		Pts:    s.pts.State(),
		Limit:  s.diffLim,
	})
	s.chDiffSem.release()
	if err != nil {
		if tgerr.Is(err, "CHANNEL_PRIVATE") {
			s.metrics.difference(ctx, kindChannel, diffInaccessible, start)
			return s.handleInaccessible(ctx)
		}
		s.metrics.difference(ctx, kindChannel, diffError, start)
		return errors.Wrap(err, "get channel difference")
	}
	result := channelDiffResult(diff)
	s.metrics.difference(ctx, kindChannel, result, start)
	if s.canceled(ctx) {
		// The response raced cancellation: do not apply it.
		return ctx.Err()
	}
	s.log.Debug(ctx, "Channel difference received",
		log.String("result", result),
		log.Duration("duration", time.Since(start)),
	)

	switch diff := diff.(type) {
	case *tg.UpdatesChannelDifference:
		if err := s.applyDifference(ctx, diff); err != nil {
			return err
		}
		if !diff.Final {
			return s.getDifference(ctx, "channel-difference-not-final")
		}
		// Only a final difference schedules polling.
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		return nil

	case *tg.UpdatesChannelDifferenceEmpty:
		if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, diff.Pts); err != nil {
			s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
			if s.stopOnApplyError {
				return &applyError{err}
			}
		}

		s.pts.SetState(diff.Pts, "updates.channelDifferenceEmpty")
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		return nil

	case *tg.UpdatesChannelDifferenceTooLong:
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		gap := ChannelGap{
			ChannelID:  s.channelID,
			AccessHash: s.accessHash,
			LocalPts:   s.pts.State(),
		}
		if d, ok := diff.Dialog.(*tg.Dialog); ok {
			gap.TopMessage = d.TopMessage
		}

		// The server returns the latest channel messages, which may include
		// messages already delivered before the gap: they are passed to
		// OnChannelGap instead of Handler, so they are never seen as new.
		gap.Messages = diff.Messages
		gap.Users = diff.Users
		gap.Chats = diff.Chats

		dialog := "ok"
		remotePts, err := getDialogPts(diff.Dialog)
		if err != nil {
			dialog = "invalid"
			s.log.Warn(ctx, "UpdatesChannelDifferenceTooLong invalid Dialog", log.Error(err))
		} else {
			// Same guard as TDLib: pts must advance, or a non-final
			// difference would be requested again for the same pts.
			if remotePts <= gap.LocalPts {
				dialog = "stale"
				s.log.Warn(ctx, "UpdatesChannelDifferenceTooLong pts did not advance",
					log.Int("local_pts", gap.LocalPts),
					log.Int("remote_pts", remotePts),
				)
				remotePts = gap.LocalPts + 1
			}
			gap.RemotePts = remotePts
			if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, remotePts); err != nil {
				s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
				if s.stopOnApplyError {
					return &applyError{err}
				}
			}

			s.pts.SetState(remotePts, "updates.channelDifferenceTooLong dialog new pts")
		}

		s.metrics.channelTooLong(ctx, dialog, gap)
		s.log.Info(ctx, "Channel difference too long, reporting gap",
			log.Int("local_pts", gap.LocalPts),
			log.Int("remote_pts", gap.RemotePts),
			log.Int("top_message", gap.TopMessage),
			log.Int("messages", len(gap.Messages)),
			log.Bool("final", diff.Final),
			log.String("dialog", dialog),
		)
		s.onTooLong(s.channelID)
		if s.onGap != nil {
			s.onGap(gap)
		}

		if !diff.Final && err == nil {
			return s.getDifference(ctx, "channel-difference-too-long-not-final")
		}
		return nil

	default:
		return errors.Errorf("unexpected channel diff type: %T", diff)
	}
}

// applyDifference applies the contents of updates.channelDifference and sets
// the channel pts to diff.Pts.
//
// Other updates of this channel are dispatched directly by the worker, without
// the pts gap check, like TDLib process_get_channel_difference_updates. They
// used to be sent to *internalState, which routed them back to this worker
// only after the pts was set to diff.Pts, so the gap check dropped them as
// outdated.
//
// Updates are dispatched in TDLib order: updateMessageID, new messages, then
// the other updates of this channel. Updates that belong to another sequence
// (common pts/qts or another channel) are not expected here; they are handed to
// *internalState to be handled as usual.
func (s *channelState) applyDifference(ctx context.Context, diff *tg.UpdatesChannelDifference) error {
	var (
		updates = make([]tg.UpdateClass, 0, len(diff.NewMessages)+len(diff.OtherUpdates))
		rest    = make([]tg.UpdateClass, 0, len(diff.OtherUpdates))
		foreign []tg.UpdateClass
	)
	for _, u := range diff.OtherUpdates {
		if _, ok := u.(*tg.UpdateMessageID); ok {
			updates = append(updates, u)
			continue
		}
		if long, ok := u.(*tg.UpdateChannelTooLong); ok && long.ChannelID == s.channelID {
			// This difference is the recovery itself.
			s.log.Debug(ctx, "Ignoring updateChannelTooLong in channel difference")
			continue
		}
		if s.foreignUpdate(u) {
			foreign = append(foreign, u)
			continue
		}
		rest = append(rest, u)
	}
	updates = append(updates, msgsToUpdates(diff.NewMessages, true)...)
	sortUpdatesByPts(rest)
	updates = append(updates, rest...)

	if len(updates) > 0 {
		if err := s.dispatch(ctx, &tg.Updates{
			Updates: updates,
			Users:   diff.Users,
			Chats:   diff.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle updates error", log.Error(err))
			if s.stopOnApplyError {
				// Do not set pts: the difference is requested again.
				return &applyError{err}
			}
		}
	}

	// Hand updates of other sequences to *internalState. The batch is sent even
	// if there are none, so that *internalState still records the access
	// hashes of the response entities, as it did when every other update was
	// routed through it.
	if len(diff.OtherUpdates) > 0 {
		if len(foreign) > 0 {
			s.log.Debug(ctx, "Channel difference has updates of other sequences",
				log.Int("count", len(foreign)),
			)
		}
		if err := s.sendOut(ctx, tracedUpdate{
			span: trace.SpanContextFromContext(ctx),
			update: &tg.Updates{
				Updates: foreign,
				Users:   diff.Users,
				Chats:   diff.Chats,
			},
		}); err != nil {
			return err
		}
	}

	// Do not advance pts if the context is done: the next difference
	// re-delivers what may not have been handled.
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "apply channel difference")
	}

	if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, diff.Pts); err != nil {
		s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
		if s.stopOnApplyError {
			return &applyError{err}
		}
	}

	s.pts.SetState(diff.Pts, "updates.channelDifference")
	// Pending updates up to the new pts were delivered by the difference.
	if err := s.pts.resyncPending(ctx); err != nil {
		return errors.Wrap(err, "apply pending")
	}
	return nil
}

// foreignUpdate reports whether u belongs to a sequence other than the pts of
// this channel: the common pts or qts, or another channel. updatePtsChanged
// belongs to the common pts: *internalState refetches the common difference.
func (s *channelState) foreignUpdate(u tg.UpdateClass) bool {
	if _, ok := u.(*tg.UpdatePtsChanged); ok {
		return true
	}
	if channelID, _, _, ok, _ := tg.IsChannelPtsUpdate(u); ok {
		return channelID != s.channelID
	}
	if _, _, ok := tg.IsPtsUpdate(u); ok {
		return true
	}
	if _, ok := tg.IsQtsUpdate(u); ok {
		return true
	}
	return false
}

// sendOut hands off updates to the *internalState worker via s.out.
//
// While waiting for the worker to accept the update, it keeps draining its own
// input queue. Otherwise the worker may block pushing into s.updates while this
// goroutine blocks pushing into s.out, producing a deadlock (both queues are
// bounded and the worker is the only reader of s.out).
//
// Drained updates are ignored on purpose: they are restored by a future
// getChannelDifference call.
func (s *channelState) sendOut(ctx context.Context, out tracedUpdate) error {
	for {
		select {
		case s.out <- out:
			return nil
		case u, ok := <-s.updates:
			if !ok {
				continue
			}
			s.log.Debug(ctx, "Ignoring update while sending channel difference",
				log.Any("update", u.update),
			)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// channelDiffResult returns the "result" metric attribute of diff.
func channelDiffResult(diff tg.UpdatesChannelDifferenceClass) string {
	switch diff.(type) {
	case *tg.UpdatesChannelDifferenceEmpty:
		return diffEmpty
	case *tg.UpdatesChannelDifferenceTooLong:
		return diffTooLong
	default:
		return diffFull
	}
}

// getDifferenceLogger calls getDifference, logging any error. It returns true
// if the worker should stop: with a nil error if the channel became
// inaccessible, or with the error that must stop the manager
// (Config.StopOnApplyError).
func (s *channelState) getDifferenceLogger(ctx context.Context, reason string) (stop bool, _ error) {
	if err := s.getDifference(ctx, reason); err != nil {
		if errors.Is(err, errChannelInaccessible) {
			return true, nil
		}
		s.log.Error(ctx, "get channel difference error", log.Error(err), log.String("reason", reason))
		if isApplyError(err) {
			return true, err
		}
	}
	return false, nil
}

// canceled mirrors internalState.canceled.
func (s *channelState) canceled(ctx context.Context) bool {
	return s.stopOnApplyError && ctx.Err() != nil
}

// handleInaccessible is called when updates.getChannelDifference reports
// CHANNEL_PRIVATE, i.e. the account has lost access to the channel (kicked,
// banned, or the channel was deleted).
//
// It notifies the inaccessible hook and signals *internalState to remove this
// channel from tracking, then returns errChannelInaccessible so the Run loop
// stops the worker.
func (s *channelState) handleInaccessible(ctx context.Context) error {
	s.log.Info(ctx, "Channel is inaccessible, removing from updates manager")
	s.metrics.channelInaccessible(ctx)
	s.onInaccessible(s.channelID)
	select {
	case s.removeChannel <- s.channelID:
	case <-ctx.Done():
	}
	return errChannelInaccessible
}

func (s *channelState) resetIdleTimer() {
	if len(s.idleTimeout.C) > 0 {
		<-s.idleTimeout.C
	}

	_ = s.idleTimeout.Reset(idleTimeout)
}
