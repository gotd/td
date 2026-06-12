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
	channelID      int64
	accessHash     int64
	selfID         int64
	diffLim        int
	client         API
	storage        StateStorage
	log            log.Helper
	tracer         trace.Tracer
	handler        telegram.UpdateHandler
	onTooLong      func(channelID int64)
	onInaccessible func(channelID int64)
	// removeChannel signals *internalState to drop this channel from tracking.
	removeChannel chan<- int64
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
	Handler               telegram.UpdateHandler
	OnChannelTooLong      func(channelID int64)
	OnChannelInaccessible func(channelID int64)
	RemoveChannel         chan<- int64
	Logger                log.Logger
	Tracer                trace.Tracer
}

func newChannelState(cfg channelStateConfig) *channelState {
	state := &channelState{
		updates: make(chan channelUpdate, 10),
		out:     cfg.Out,

		idleTimeout: time.NewTimer(idleTimeout),
		done:        make(chan struct{}),

		channelID:      cfg.ChannelID,
		accessHash:     cfg.AccessHash,
		selfID:         cfg.SelfID,
		diffLim:        cfg.DiffLimit,
		client:         cfg.RawClient,
		storage:        cfg.Storage,
		log:            log.For(cfg.Logger),
		handler:        cfg.Handler,
		onTooLong:      cfg.OnChannelTooLong,
		onInaccessible: cfg.OnChannelInaccessible,
		removeChannel:  cfg.RemoveChannel,
		tracer:         cfg.Tracer,
	}

	state.pts = newSequenceBox(sequenceConfig{
		InitialState: cfg.InitialPts,
		Apply:        state.applyPts,
		Logger:       log.Named(cfg.Logger, "pts"),
		Tracer:       cfg.Tracer,
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
	if err := s.getDifference(ctx); err != nil {
		if errors.Is(err, errChannelInaccessible) {
			return nil
		}
		s.log.Error(ctx, "Failed to subscribe to channel updates", log.Error(err))
	}

	for {
		select {
		case u := <-s.updates:
			ctx := trace.ContextWithSpanContext(ctx, u.span)
			if err := s.handleUpdate(ctx, u.update, u.entities); err != nil {
				if errors.Is(err, errChannelInaccessible) {
					return nil
				}
				s.log.Error(ctx, "Handle update error", log.Error(err))
			}
		case <-s.pts.gapTimeout.C:
			s.log.Debug(ctx, "Gap timeout")
			if s.getDifferenceLogger(ctx) {
				return nil
			}
		case <-ctx.Done():
			if len(s.pts.pending) > 0 {
				// This will probably fail.
				s.getDifferenceLogger(ctx)
			}
			return ctx.Err()
		case <-s.idleTimeout.C:
			s.log.Debug(ctx, "Idle timeout")
			s.resetIdleTimer()
			if s.getDifferenceLogger(ctx) {
				return nil
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

func (s *channelState) handleTooLong(ctx context.Context, long *tg.UpdateChannelTooLong) error {
	ctx, span := s.tracer.Start(ctx, "channelState.handleTooLong")
	defer span.End()

	remotePts, ok := long.GetPts()
	if !ok {
		s.log.Warn(ctx, "Got UpdateChannelTooLong without pts field")
		return s.getDifference(ctx)
	}

	// Note: we still can fetch latest diffLim updates.
	// Should we do?
	if remotePts-s.pts.State() > s.diffLim {
		s.onTooLong(s.channelID)
		return nil
	}

	return s.getDifference(ctx)
}

func (s *channelState) applyPts(ctx context.Context, state int, updates []update) error {
	ctx, span := s.tracer.Start(ctx, "channelState.applyPts")
	defer span.End()

	var (
		converted []tg.UpdateClass
		ents      entities
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Entities)
	}

	if err := s.handler.Handle(ctx, &tg.Updates{
		Updates: converted,
		Users:   ents.Users,
		Chats:   ents.Chats,
	}); err != nil {
		s.log.Error(ctx, "Handle update error", log.Error(err))
		return nil
	}

	if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, state); err != nil {
		s.log.Error(ctx, "SetChannelPts error", log.Error(err))
	}

	return nil
}

func (s *channelState) getDifference(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "channelState.getDifference")
	defer span.End()
	s.pts.gaps.Clear()

	s.log.Debug(ctx, "Getting difference")

	if now := time.Now(); now.Before(s.diffTimeout) {
		dur := s.diffTimeout.Sub(now)
		s.log.Debug(ctx, "GetChannelDifference timeout", log.Duration("duration", dur))
		if err := func() error {
			afterC := time.After(dur)
			for {
				select {
				case <-afterC:
					return nil
				case u, ok := <-s.updates:
					if !ok {
						continue
					}

					// Ignoring updates to prevent *internalState worker from blocking.
					// All ignored updates should be restored by future getChannelDifference call.
					// At least I hope so...
					s.log.Debug(ctx, "Ignoring update due to getChannelDifference timeout", log.Any("update", u.update))
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}(); err != nil {
			return err
		}
	}

	diff, err := s.client.UpdatesGetChannelDifference(ctx, &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{
			ChannelID:  s.channelID,
			AccessHash: s.accessHash,
		},
		Filter: &tg.ChannelMessagesFilterEmpty{},
		Pts:    s.pts.State(),
		Limit:  s.diffLim,
	})
	if err != nil {
		if tgerr.Is(err, "CHANNEL_PRIVATE") {
			return s.handleInaccessible(ctx)
		}
		return errors.Wrap(err, "get channel difference")
	}

	switch diff := diff.(type) {
	case *tg.UpdatesChannelDifference:
		if len(diff.OtherUpdates) > 0 {
			if err := s.sendOut(ctx, tracedUpdate{
				span: trace.SpanContextFromContext(ctx),
				update: &tg.Updates{
					Updates: diff.OtherUpdates,
					Users:   diff.Users,
					Chats:   diff.Chats,
				},
			}); err != nil {
				return err
			}
		}

		if len(diff.NewMessages) > 0 {
			if err := s.handler.Handle(ctx, &tg.Updates{
				Updates: msgsToUpdates(diff.NewMessages, true),
				Users:   diff.Users,
				Chats:   diff.Chats,
			}); err != nil {
				s.log.Error(ctx, "Handle updates error", log.Error(err))
			}
		}

		if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, diff.Pts); err != nil {
			s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
		}

		s.pts.SetState(diff.Pts, "updates.channelDifference")
		if seconds, ok := diff.GetTimeout(); ok {
			s.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		if !diff.Final {
			return s.getDifference(ctx)
		}

		return nil

	case *tg.UpdatesChannelDifferenceEmpty:
		if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, diff.Pts); err != nil {
			s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
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

		remotePts, err := getDialogPts(diff.Dialog)
		if err != nil {
			s.log.Warn(ctx, "UpdatesChannelDifferenceTooLong invalid Dialog", log.Error(err))
		} else {
			if err := s.storage.SetChannelPts(ctx, s.selfID, s.channelID, remotePts); err != nil {
				s.log.Warn(ctx, "SetChannelPts error", log.Error(err))
			}

			s.pts.SetState(remotePts, "updates.channelDifferenceTooLong dialog new pts")
		}

		s.onTooLong(s.channelID)
		return nil

	default:
		return errors.Errorf("unexpected channel diff type: %T", diff)
	}
}

// sendOut hands off updates to the *internalState worker via s.out.
//
// While waiting for the worker to accept the update, it keeps draining its own
// input queue. Otherwise the worker may block pushing into s.updates while this
// goroutine blocks pushing into s.out, producing a deadlock (both queues are
// bounded and the worker is the only reader of s.out).
//
// Drained updates are ignored on purpose: they are restored by a future
// getChannelDifference call, same as in the diffTimeout wait above.
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

// getDifferenceLogger calls getDifference, logging any error. It returns true
// if the channel became inaccessible and the worker should stop.
func (s *channelState) getDifferenceLogger(ctx context.Context) (stop bool) {
	if err := s.getDifference(ctx); err != nil {
		if errors.Is(err, errChannelInaccessible) {
			return true
		}
		s.log.Error(ctx, "get channel difference error", log.Error(err))
	}
	return false
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
