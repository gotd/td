package updates

import (
	"context"

	"go.uber.org/zap"

	"github.com/nnqq/td/tg"
)

func (s *state) applySeq(ctx context.Context, state int, updates []update) error {
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

	if err := s.storage.SetSeq(s.selfID, state); err != nil {
		s.log.Error("SetSeq error", zap.Error(err))
	}

	if recoverState {
		return s.getDifference()
	}

	return nil
}

func (s *state) applyCombined(ctx context.Context, comb *tg.UpdatesCombined) (ptsChanged bool, err error) {
	var (
		ents = entities{
			Users: comb.Users,
			Chats: comb.Chats,
		}
		others []tg.UpdateClass
	)

	for _, u := range comb.Updates {
		switch u := u.(type) {
		case *tg.UpdatePtsChanged:
			ptsChanged = true
			continue
		case *tg.UpdateChannelTooLong:
			channelState, ok := s.channels[u.ChannelID]
			if !ok {
				s.log.Debug("ChannelTooLong for channel that is not in the state, update ignored", zap.Int64("channel_id", u.ChannelID))
				continue
			}

			channelState.PushUpdate(channelUpdate{u, ctx, ents})
			continue
		}

		if pts, ptsCount, ok := isCommonPtsUpdate(u); ok {
			if err := s.handlePts(pts, ptsCount, u, ents); err != nil {
				return false, err
			}

			continue
		}

		if channelID, pts, ptsCount, ok, err := isChannelPtsUpdate(u); ok {
			if err != nil {
				s.log.Debug("Invalid channel update", zap.Error(err), zap.Any("update", u))
				continue
			}

			s.handleChannel(channelID, comb.Date, pts, ptsCount, channelUpdate{
				update: u,
				ctx:    ctx,
				ents:   ents,
			})
			continue
		}

		if qts, ok := isCommonQtsUpdate(u); ok {
			if err := s.handleQts(qts, u, ents); err != nil {
				return false, err
			}

			continue
		}

		others = append(others, u)
	}

	if len(others) > 0 {
		if err := s.handler.Handle(s.ctx, &tg.Updates{
			Updates: others,
			Users:   ents.Users,
			Chats:   ents.Chats,
		}); err != nil {
			s.log.Error("Handle updates error", zap.Error(err))
		}
	}

	setDate, setSeq := comb.Date > s.date, comb.Seq > 0
	switch {
	case setDate && setSeq:
		if err := s.storage.SetDateSeq(s.selfID, comb.Date, comb.Seq); err != nil {
			s.log.Error("SetDateSeq error", zap.Error(err))
		}

		s.date = comb.Date
		s.seq.SetState(comb.Seq)
	case setDate:
		if err := s.storage.SetDate(s.selfID, comb.Date); err != nil {
			s.log.Error("SetDate error", zap.Error(err))
		}
		s.date = comb.Date
	case setSeq:
		if err := s.storage.SetSeq(s.selfID, comb.Seq); err != nil {
			s.log.Error("SetSeq error", zap.Error(err))
		}
		s.seq.SetState(comb.Seq)
	}

	return ptsChanged, nil
}

// nolint:dupl
func (s *state) applyPts(ctx context.Context, state int, updates []update) error {
	var (
		converted []tg.UpdateClass
		ents      entities
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Ents)
	}

	if err := s.handler.Handle(s.ctx, &tg.Updates{
		Updates: converted,
		Users:   ents.Users,
		Chats:   ents.Chats,
	}); err != nil {
		s.log.Error("Handle updates error", zap.Error(err))
	}

	if err := s.storage.SetPts(s.selfID, state); err != nil {
		s.log.Error("SetPts error", zap.Error(err))
	}

	return nil
}

// nolint:dupl
func (s *state) applyQts(ctx context.Context, state int, updates []update) error {
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
		s.log.Error("Handle updates error", zap.Error(err))
	}

	if err := s.storage.SetQts(s.selfID, state); err != nil {
		s.log.Error("SetQts error", zap.Error(err))
	}

	return nil
}
