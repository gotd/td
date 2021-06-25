package updates

import "go.uber.org/zap"

func (e *Engine) initCommonBoxes(state State) {
	recoverState := func() {
		if err := e.recoverState(); err != nil {
			e.log.Warn("Recover state error", zap.Error(err))
		}
	}

	e.setDate(state.Date)
	e.pts = newSequenceBox(sequenceConfig{
		InitialState: state.Pts,
		Apply:        e.applyPts,
		OnGap:        recoverState,
		Logger:       e.log.Named("pts"),
	})
	e.qts = newSequenceBox(sequenceConfig{
		InitialState: state.Qts,
		Apply:        e.applyQts,
		OnGap:        recoverState,
		Logger:       e.log.Named("qts"),
	})
	e.seq = newSequenceBox(sequenceConfig{
		InitialState: state.Seq,
		Apply:        e.applySeq,
		OnGap:        recoverState,
		Logger:       e.log.Named("seq"),
	})

	e.seq.run()
	e.pts.run()
	e.qts.run()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-e.done:
				return

			case <-e.recoverGap:
				recoverState()

			case <-e.idleTimeout.C:
				e.log.Debug("Idle timeout, recovering state")
				_ = e.idleTimeout.Reset(idleTimeout)
				recoverState()
			}
		}
	}()
}

func (e *Engine) stopCommonBoxes() {
	close(e.done)
	e.wg.Wait()

	e.seq.stop()
	e.pts.stop()
	e.qts.stop()
}
