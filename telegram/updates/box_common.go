package updates

func (e *Engine) initCommonBoxes(state State) {
	recover := func() {
		if err := e.recoverState(); err != nil {
			e.echan <- err
		}
	}

	e.setDate(state.Date)
	e.pts = newSequenceBox(sequenceConfig{
		InitialState: state.Pts,
		Apply:        e.applyPts,
		OnGap:        recover,
		Logger:       e.log.Named("pts"),
	})
	e.qts = newSequenceBox(sequenceConfig{
		InitialState: state.Qts,
		Apply:        e.applyQts,
		OnGap:        recover,
		Logger:       e.log.Named("qts"),
	})
	e.seq = newSequenceBox(sequenceConfig{
		InitialState: state.Seq,
		Apply:        e.applySeq,
		OnGap:        recover,
		Logger:       e.log.Named("seq"),
	})

	e.pts.run()
	e.qts.run()
	e.seq.run()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-e.workers:
				return

			case <-e.recoverGap:
				if err := e.recoverState(); err != nil {
					e.echan <- err
				}

			case <-e.idleTimeout.C:
				e.log.Info("Idle timeout, recovering state")
				_ = e.idleTimeout.Reset(idleTimeout)
				if err := e.recoverState(); err != nil {
					e.echan <- err
				}
			}
		}
	}()
}
