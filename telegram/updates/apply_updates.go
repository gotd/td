package updates

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (e *Engine) applySeq(state int, updates []update) error {
	recoverState := false
	for _, u := range updates {
		ptsChanged, err := e.applyCombined(u.Value.(*tg.UpdatesCombined))
		if err != nil {
			return err
		}

		if ptsChanged {
			recoverState = true
		}
	}

	if recoverState {
		e.recoverGap <- struct{}{}
	}
	return nil
}

// nolint:gocognit
func (e *Engine) applyCombined(comb *tg.UpdatesCombined) (ptsChanged bool, err error) {
	var (
		ents   = NewEntities().fromUpdates(comb)
		others []tg.UpdateClass
	)

	for _, u := range comb.Updates {
		switch u := u.(type) {
		case *tg.UpdatePtsChanged:
			ptsChanged = true
			continue
		case *tg.UpdateChannelTooLong:
			e.handleChannelTooLong(comb.Date, u)
			continue
		}

		if pts, ptsCount, ok := isCommonPtsUpdate(u); ok {
			if err := e.handlePts(pts, ptsCount, u, ents); err != nil {
				return false, err
			}

			continue
		}

		if channelID, pts, ptsCount, ok, err := isChannelPtsUpdate(u); ok {
			if err != nil {
				e.log.Warn("Invalid channel update", zap.Error(err))
				continue
			}

			if err := e.handleChannel(channelID, comb.Date, pts, ptsCount, u, ents); err != nil {
				return false, err
			}

			continue
		}

		if qts, ok := isCommonQtsUpdate(u); ok {
			if err := e.handleQts(qts, u, ents); err != nil {
				return false, err
			}

			continue
		}

		others = append(others, u)
	}

	if len(others) > 0 {
		if err := e.handler.HandleUpdates(ents, others); err != nil {
			return false, xerrors.Errorf("handle updates: %w", err)
		}
	}

	if comb.Seq > 0 {
		if err := e.storage.SetSeq(comb.Seq); err != nil {
			return false, err
		}
	}

	if comb.Date > 0 {
		e.dateMux.Lock()
		defer e.dateMux.Unlock()

		if comb.Date > e.date {
			if err := e.storage.SetDate(comb.Date); err != nil {
				return false, err
			}

			e.date = comb.Date
		}
	}

	return ptsChanged, nil
}

func (e *Engine) applyPts(state int, updates []update) error {
	var (
		converted []tg.UpdateClass
		ents      = NewEntities()
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Ents)
	}

	if err := e.handler.HandleUpdates(ents, converted); err != nil {
		return err
	}

	return e.storage.SetPts(state)
}

func (e *Engine) applyQts(state int, updates []update) error {
	var (
		converted []tg.UpdateClass
		ents      = NewEntities()
	)

	for _, update := range updates {
		converted = append(converted, update.Value.(tg.UpdateClass))
		ents.Merge(update.Ents)
	}

	if err := e.handler.HandleUpdates(ents, converted); err != nil {
		return err
	}

	return e.storage.SetQts(state)
}

func (e *Engine) applyChannel(channelID int) func(state int, updates []update) error {
	return func(state int, updates []update) error {
		var (
			converted []tg.UpdateClass
			ents      = NewEntities()
		)

		for _, update := range updates {
			converted = append(converted, update.Value.(tg.UpdateClass))
			ents.Merge(update.Ents)
		}

		if err := e.handler.HandleUpdates(ents, converted); err != nil {
			return err
		}

		return e.storage.SetChannelPts(channelID, state)
	}
}
