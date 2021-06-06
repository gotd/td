package updates

import (
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type channelState struct {
	pts         *sequenceBox
	recoverGap  chan struct{}
	recovering  atomic.Bool
	idleTimeout *time.Timer
	diffTimeout time.Time
}

func (e *Engine) createChannelState(channelID, initialPts int) *channelState {
	state := new(channelState)
	state.recoverGap = make(chan struct{}, 2)
	state.idleTimeout = time.NewTimer(idleTimeout)
	state.pts = newSequenceBox(sequenceConfig{
		InitialState: initialPts,
		Apply:        e.applyChannel(channelID),
		OnGap:        func() { state.recoverGap <- struct{}{} },
		Logger:       e.log.Named("channel_pts").With(zap.Int("channel_id", channelID)),
	})
	state.pts.run()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-e.workers:
				return

			case <-state.recoverGap:
				if err := e.recoverChannelState(channelID, state); err != nil {
					e.echan <- err
				}

			case <-state.idleTimeout.C:
				state.pts.log.Info("Idle timeout, recovering state")
				_ = state.idleTimeout.Reset(idleTimeout)
				if err := e.recoverChannelState(channelID, state); err != nil {
					e.echan <- err
				}
			}
		}
	}()

	return state
}
