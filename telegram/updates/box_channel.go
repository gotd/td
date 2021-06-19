package updates

import (
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type channelState struct {
	channelID   int
	pts         *sequenceBox
	recoverGap  chan struct{}
	recovering  atomic.Bool
	idleTimeout *time.Timer
	diffTimeout time.Time

	e    *Engine
	wg   sync.WaitGroup
	done chan struct{}
}

func (e *Engine) createChannelState(channelID, initialPts int) *channelState {
	recoverGap := make(chan struct{}, 2)
	state := &channelState{
		channelID: channelID,
		pts: newSequenceBox(sequenceConfig{
			InitialState: initialPts,
			Apply:        e.applyChannel(channelID),
			OnGap:        func() { recoverGap <- struct{}{} },
			Logger:       e.log.Named("channel_pts").With(zap.Int("channel_id", channelID)),
		}),
		recoverGap:  recoverGap,
		idleTimeout: time.NewTimer(idleTimeout),

		e:    e,
		done: make(chan struct{}),
	}

	state.run()
	recoverGap <- struct{}{}
	return state
}

func (s *channelState) run() {
	s.wg.Add(1)
	s.pts.run()

	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.done:
				return

			case <-s.recoverGap:
				s.recoverState()

			case <-s.idleTimeout.C:
				s.pts.log.Debug("Idle timeout, recovering state")
				_ = s.idleTimeout.Reset(idleTimeout)
				s.recoverState()
			}
		}
	}()
}

func (s *channelState) stop() {
	close(s.done)
	s.wg.Wait()
	s.pts.stop()
	_ = s.idleTimeout.Stop()
}

func (s *channelState) recoverState() {
	if err := s.e.recoverChannelState(s); err != nil {
		s.e.log.Warn("Recover channel state error",
			zap.Int("channel_id", s.channelID),
			zap.Error(err),
		)
	}
}

func (e *Engine) removeChannelState(channelID int) {
	e.chanMux.Lock()
	defer e.chanMux.Unlock()
	state, ok := e.channels[channelID]
	if !ok {
		return
	}

	delete(e.channels, channelID)
	state.stop()
}
