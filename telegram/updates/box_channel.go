package updates

import (
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type channelState struct {
	pts         *sequenceBox
	recovering  atomic.Bool
	idleTimeout *time.Timer
	diffTimeout time.Time
}

func (e *Engine) createChannelState(channelID, initialPts int) *channelState {
	state := new(channelState)
	state.pts = newSequenceBox(sequenceConfig{
		InitialState: initialPts,
		Apply:        e.applyChannel(channelID),
		OnGap:        e.channelGapHandler(channelID, state),
		Logger:       e.log.Named("channel_pts").With(zap.Int("channel_id", channelID)),
	})
	state.idleTimeout = time.NewTimer(idleTimeout)

	go state.pts.run(e.ctx)

	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return

			case <-state.idleTimeout.C:
				state.pts.log.Info("Idle timeout, recovering state")
				go func() {
					if err := e.recoverChannelState(channelID, state); err != nil {
						e.echan <- err
					}
				}()
			}
		}
	}()

	return state
}

func (e *Engine) initChannelBoxes() error {
	if err := e.storage.Channels(func(channelID, pts int) {
		e.chanMux.Lock()
		e.channels[channelID] = e.createChannelState(channelID, pts)
		e.chanMux.Unlock()
	}); err != nil {
		return xerrors.Errorf("restore local channels state: %w", err)
	}

	for channelID, state := range e.channels {
		_ = e.recoverChannelState(channelID, state)
	}

	return nil
}

func (e *Engine) channelGapHandler(channelID int, ch *channelState) func(state gapState) {
	return func(state gapState) {
		switch state {
		case gapInit:
			e.wg.Add(1)
		case gapResolved:
			e.wg.Done()
		case gapRecover:
			go func() {
				defer e.wg.Done()

				if err := e.recoverChannelState(channelID, ch); err != nil {
					e.echan <- err
				}
			}()
		}
	}
}
