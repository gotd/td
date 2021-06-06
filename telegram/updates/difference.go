package updates

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (e *Engine) recoverState() error {
	if !e.recovering.CAS(false, true) {
		return nil
	}
	defer e.recovering.Store(false)

	e.pts.EnableRecoverMode()
	e.qts.EnableRecoverMode()
	e.seq.EnableRecoverMode()

	defer func() {
		_ = e.pts.ExtractBuffer()
		_ = e.qts.ExtractBuffer()
		_ = e.seq.ExtractBuffer()

		e.pts.DisableRecoverMode()
		e.qts.DisableRecoverMode()
		e.seq.DisableRecoverMode()
	}()

	e.log.Info("Recovering state")
	if err := e.getDifference(); err != nil {
		e.log.Error("Recover state error", zap.Error(err))
		return xerrors.Errorf("getDifference: %w", err)
	}

	e.log.Info("State recovered")
	return nil
}

func (e *Engine) recoverChannelState(channelID int, state *channelState) error {
	if !state.recovering.CAS(false, true) {
		return nil
	}
	defer state.recovering.Store(false)

	log := e.log.With(zap.Int("channel_id", channelID))
	accessHash, ok := e.getChannelAccessHash(channelID, 0)
	if !ok {
		log.Warn("Cannot recover state due to missing access hash.")
		return nil
	}

	state.pts.EnableRecoverMode()
	defer func() {
		_ = state.pts.ExtractBuffer()
		state.pts.DisableRecoverMode()
	}()

	log.Info("Recovering state")
	if err := e.getChannelDifference(channelID, accessHash, state); err != nil {
		log.Error("Recover channel state error", zap.Error(err))
		return xerrors.Errorf("getChannelDifference(id: %d): %w", channelID, err)
	}

	log.Info("State recovered")
	return nil
}

func (e *Engine) getDifference() error {
	setState := func(state tg.UpdatesState) error {
		if err := e.storage.SetState((State{}).fromRemote(&state)); err != nil {
			return err
		}

		e.pts.SetState(state.Pts)
		e.qts.SetState(state.Qts)
		e.seq.SetState(state.Seq)
		e.setDate(state.Date)
		return nil
	}

	diff, err := e.raw.UpdatesGetDifference(e.ctx, &tg.UpdatesGetDifferenceRequest{
		Pts:  e.pts.State(),
		Qts:  e.qts.State(),
		Date: e.getDate(),
	})
	if err != nil {
		return xerrors.Errorf("get difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		e.saveChannelHashes("UpdatesDifference", diff.Chats)
		if len(diff.OtherUpdates) > 0 {
			if err := e.handleUpdates(&tg.Updates{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
				Date:    diff.State.Date,
			}); err != nil {
				return err
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			d := DiffUpdate{
				NewMessages:          diff.NewMessages,
				NewEncryptedMessages: diff.NewEncryptedMessages,
				Users:                diff.Users,
				Chats:                diff.Chats,
			}

			if err := e.handler.HandleDiff(d); err != nil {
				return err
			}
		}

		if err := setState(diff.State); err != nil {
			return err
		}

		return nil

	// No events.
	case *tg.UpdatesDifferenceEmpty:
		if err := e.storage.SetSeq(diff.Seq); err != nil {
			return err
		}
		if err := e.storage.SetDate(diff.Date); err != nil {
			return err
		}

		e.setDate(diff.Date)
		e.seq.SetState(diff.Seq)

		return nil

	// Incomplete list of occurred events.
	case *tg.UpdatesDifferenceSlice:
		e.saveChannelHashes("UpdatesDifferenceSlice", diff.Chats)

		if len(diff.OtherUpdates) > 0 {
			if err := e.handleUpdates(&tg.Updates{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
				Date:    diff.IntermediateState.Date,
			}); err != nil {
				return err
			}
		}

		if len(diff.NewMessages) > 0 || len(diff.NewEncryptedMessages) > 0 {
			d := DiffUpdate{
				NewMessages:          diff.NewMessages,
				NewEncryptedMessages: diff.NewEncryptedMessages,
				Users:                diff.Users,
				Chats:                diff.Chats,
			}

			if err := e.handler.HandleDiff(d); err != nil {
				return err
			}
		}

		if err := setState(diff.IntermediateState); err != nil {
			return err
		}

		return e.getDifference()

	// The difference is too long, and the specified state must be used to refetch updates.
	case *tg.UpdatesDifferenceTooLong:
		if err := e.storage.SetPts(diff.Pts); err != nil {
			return err
		}

		e.pts.SetState(diff.Pts)
		return e.getDifference()

	default:
		return xerrors.Errorf("unexpected diff type: %T", diff)
	}
}

func (e *Engine) getChannelDifference(channelID int, accessHash int64, state *channelState) error {
	if now := time.Now(); now.Before(state.diffTimeout) {
		time.Sleep(state.diffTimeout.Sub(now))
	}

	diff, err := e.raw.UpdatesGetChannelDifference(e.ctx, &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
		Filter: &tg.ChannelMessagesFilterEmpty{},
		Pts:    state.pts.State(),
		Limit:  e.diffLim,
	})
	if err != nil {
		return xerrors.Errorf("get channel difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesChannelDifference:
		e.saveChannelHashes("UpdatesChannelDifference", diff.Chats)

		if len(diff.OtherUpdates) > 0 {
			if err := e.handleUpdates(&tg.Updates{
				Updates: diff.OtherUpdates,
				Users:   diff.Users,
				Chats:   diff.Chats,
			}); err != nil {
				return err
			}
		}

		if len(diff.NewMessages) > 0 {
			d := DiffUpdate{
				NewMessages: diff.NewMessages,
				Users:       diff.Users,
				Chats:       diff.Chats,
			}

			if err := e.handler.HandleDiff(d); err != nil {
				return err
			}
		}

		if err := e.storage.SetChannelPts(channelID, diff.Pts); err != nil {
			return err
		}

		state.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			state.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		if !diff.Final {
			return e.getChannelDifference(channelID, accessHash, state)
		}

		return nil

	case *tg.UpdatesChannelDifferenceEmpty:
		if err := e.storage.SetChannelPts(channelID, diff.Pts); err != nil {
			return err
		}

		state.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			state.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		return nil

	case *tg.UpdatesChannelDifferenceTooLong:
		e.saveChannelHashes("UpdatesChannelDifferenceTooLong", diff.Chats)

		// Reset channel state.
		e.chanMux.Lock()
		delete(e.channels, channelID)
		e.chanMux.Unlock()

		e.handler.ChannelTooLong(channelID)
		if seconds, ok := diff.GetTimeout(); ok {
			state.diffTimeout = time.Now().Add(time.Second * time.Duration(seconds))
		}

		return nil

	default:
		return xerrors.Errorf("unexpected channel diff type: %T", diff)
	}
}

func (e *Engine) saveChannelHashes(source string, chats []tg.ChatClass) {
	e.hashMux.Lock()
	defer e.hashMux.Unlock()
	e.saveChannelHashesNoMux(source, chats)
}

func (e *Engine) saveChannelHashesNoMux(source string, chats []tg.ChatClass) {
	for _, c := range chats {
		switch c := c.(type) {
		case *tg.Channel:
			if hash, ok := c.GetAccessHash(); ok && !c.Min {
				if _, ok := e.channelHash[c.ID]; !ok {
					e.log.Info("New channel access hash",
						zap.Int("channel_id", c.ID),
						zap.String("channel_name", c.Username),
						zap.String("source", source),
					)
				}
				e.channelHash[c.ID] = hash
			}
		case *tg.ChannelForbidden:
			if _, ok := e.channelHash[c.ID]; !ok {
				e.log.Info("New forbidden channel access hash",
					zap.Int("channel_id", c.ID),
					zap.String("channel_title", c.Title),
					zap.String("source", source),
				)
			}
			e.channelHash[c.ID] = c.AccessHash
		}
	}
}

func (e *Engine) getChannelAccessHash(channelID, date int) (int64, bool) {
	e.hashMux.Lock()
	defer e.hashMux.Unlock()

	log := e.log.With(zap.Int("channel_id", channelID))
	accessHash, ok := e.channelHash[channelID]
	if !ok {
		if date == 0 {
			// Update have no date, fallback to global.
			date = e.getDate() - 31
		}

		diff, err := e.raw.UpdatesGetDifference(e.ctx, &tg.UpdatesGetDifferenceRequest{
			Pts:  e.pts.State(),
			Qts:  e.qts.State(),
			Date: date - 1,
		})
		if err != nil {
			log.Warn("Restore access hash error", zap.Error(err))
			return 0, false
		}

		switch diff := diff.(type) {
		case *tg.UpdatesDifference:
			e.saveChannelHashesNoMux("UpdatesDifference", diff.Chats)
		case *tg.UpdatesDifferenceSlice:
			e.saveChannelHashesNoMux("UpdatesDifferenceSlice", diff.Chats)
		}

		accessHash, ok = e.channelHash[channelID]
		if !ok {
			log.Warn("Failed to restore access hash: getDifference result does not contain expected hash")
			return 0, false
		}
	}

	return accessHash, true
}

func (e *Engine) getDate() int {
	e.dateMux.Lock()
	defer e.dateMux.Unlock()
	return e.date
}

func (e *Engine) setDate(date int) {
	e.dateMux.Lock()
	defer e.dateMux.Unlock()
	e.date = date
}
