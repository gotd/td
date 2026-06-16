package updates

import (
	"context"

	"github.com/gotd/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/tg"
)

func (s *internalState) applySeq(ctx context.Context, state int, updates []update) error {
	recoverState := false
	for _, u := range updates {
		ptsChanged, err := s.applyCombined(ctx, u.Value.(*tg.UpdatesCombined))
		if err != nil {
			return err
		}

		if ptsChanged {
			recoverState = true
		}
	}

	if err := s.storage.SetSeq(ctx, s.selfID, state); err != nil {
		s.log.Error(ctx, "SetSeq error", log.Error(err))
	}

	if recoverState {
		return s.getDifference(ctx, "seq-combined-pts-changed")
	}

	return nil
}

func (s *internalState) applyCombined(ctx context.Context, comb *tg.UpdatesCombined) (ptsChanged bool, err error) {
	ctx, span := s.tracer.Start(ctx, "internalState.applyCombined")
	defer span.End()

	var (
		ents = entities{
			Users: comb.Users,
			Chats: comb.Chats,
		}
		others []tg.UpdateClass
	)
	sortUpdatesByPts(comb.Updates)

	for _, u := range comb.Updates {
		switch u := u.(type) {
		case *tg.UpdatePtsChanged:
			ptsChanged = true
			continue
		case *tg.UpdateChannelTooLong:
			st, ok := s.channels[u.ChannelID]
			if !ok {
				s.log.Debug(ctx, "ChannelTooLong for channel that is not in the internalState, update ignored", log.Int64("channel_id", u.ChannelID))
				continue
			}
			if err := st.Push(ctx, channelUpdate{
				update:   u,
				entities: ents,
				span:     trace.SpanContextFromContext(ctx),
			}); err != nil {
				s.log.Error(ctx, "Push channel update error", log.Error(err))
			}
			continue
		}

		if pts, ptsCount, ok := tg.IsPtsUpdate(u); ok {
			if err := s.handlePts(ctx, pts, ptsCount, u, ents); err != nil {
				return false, err
			}

			continue
		}

		if channelID, pts, ptsCount, ok, err := tg.IsChannelPtsUpdate(u); ok {
			if err != nil {
				s.log.Debug(ctx, "Invalid channel update", log.Error(err), log.Any("update", u))
				continue
			}
			if err := s.handleChannel(ctx, channelID, comb.Date, pts, ptsCount, channelUpdate{
				update:   u,
				entities: ents,
				span:     trace.SpanContextFromContext(ctx),
			}); err != nil {
				s.log.Error(ctx, "Handle channel update error", log.Error(err))
			}

			continue
		}

		if qts, ok := tg.IsQtsUpdate(u); ok {
			if err := s.handleQts(ctx, qts, u, ents); err != nil {
				return false, err
			}

			continue
		}

		others = append(others, u)
	}

	if len(others) > 0 {
		if err := s.dispatch(ctx, &tg.Updates{
			Updates: others,
			Users:   ents.Users,
			Chats:   ents.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle updates error", log.Error(err))
		}
	}

	setDate, setSeq := comb.Date > s.date, comb.Seq > 0
	switch {
	case setDate && setSeq:
		if err := s.storage.SetDateSeq(ctx, s.selfID, comb.Date, comb.Seq); err != nil {
			s.log.Error(ctx, "SetDateSeq error", log.Error(err))
		}

		s.date = comb.Date
		s.seq.SetState(comb.Seq, "seq update")
	case setDate:
		if err := s.storage.SetDate(ctx, s.selfID, comb.Date); err != nil {
			s.log.Error(ctx, "SetDate error", log.Error(err))
		}
		s.date = comb.Date
	case setSeq:
		if err := s.storage.SetSeq(ctx, s.selfID, comb.Seq); err != nil {
			s.log.Error(ctx, "SetSeq error", log.Error(err))
		}
		s.seq.SetState(comb.Seq, "seq update")
	}

	return ptsChanged, nil
}

func (s *internalState) applyPts(ctx context.Context, state int, updates []update) error {
	ctx, span := s.tracer.Start(ctx, "internalState.applyPts")
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
			s.log.Error(ctx, "Handle updates error", log.Error(err))
		}
	}

	if err := s.storage.SetPts(ctx, s.selfID, state); err != nil {
		s.log.Error(ctx, "SetPts error", log.Error(err))
	}

	return nil
}

func (s *internalState) applyQts(ctx context.Context, state int, updates []update) error {
	ctx, span := s.tracer.Start(ctx, "internalState.applyQts")
	defer span.End()

	var (
		converted []tg.UpdateClass
		ents      entities
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Entities)
	}

	if err := s.dispatch(ctx, &tg.Updates{
		Updates: converted,
		Users:   ents.Users,
		Chats:   ents.Chats,
	}); err != nil {
		s.log.Error(ctx, "Handle updates error", log.Error(err))
	}

	// Don't set qts if it's 0, because it means that we are apllying gaps updates
	if state == 0 {
		return nil
	}

	if err := s.storage.SetQts(ctx, s.selfID, state); err != nil {
		s.log.Error(ctx, "SetQts error", log.Error(err))
	}

	return nil
}
