package updates

func (e *Engine) initCommonBoxes(state State) error {
	e.setDate(state.Date)
	e.pts = newSequenceBox(sequenceConfig{
		InitialState: state.Pts,
		Apply:        e.applyPts,
		OnGap:        e.commonGapHandler,
		Logger:       e.log.Named("pts"),
	})
	e.qts = newSequenceBox(sequenceConfig{
		InitialState: state.Qts,
		Apply:        e.applyQts,
		OnGap:        e.commonGapHandler,
		Logger:       e.log.Named("qts"),
	})
	e.seq = newSequenceBox(sequenceConfig{
		InitialState: state.Seq,
		Apply:        e.applySeq,
		OnGap:        e.commonGapHandler,
		Logger:       e.log.Named("seq"),
	})

	go e.pts.run(e.ctx)
	go e.qts.run(e.ctx)
	go e.seq.run(e.ctx)

	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return

			case <-e.idleTimeout.C:
				e.log.Info("Idle timeout, recovering state")
				go func() {
					if err := e.recoverState(); err != nil {
						e.echan <- err
					}
				}()
			}
		}
	}()

	return nil
}

func (e *Engine) commonGapHandler(state gapState) {
	switch state {
	case gapInit:
		e.wg.Add(1)
	case gapResolved:
		e.wg.Done()
	case gapRecover:
		go func() {
			defer e.wg.Done()

			if err := e.recoverState(); err != nil {
				e.echan <- err
			}
		}()
	}
}
