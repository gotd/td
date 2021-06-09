package updates

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// HandleUpdates handles updates.
func (e *Engine) HandleUpdates(u tg.UpdatesClass) error {
	if e.closed.Load() {
		return xerrors.Errorf("closed")
	}

	e.uchan <- u
	return nil
}

// Run engine.
func (e *Engine) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	state, err := e.getState(ctx)
	if err != nil {
		return err
	}

	if err := e.storage.Channels(func(channelID, pts int) {
		e.chanMux.Lock()
		e.channels[channelID] = e.createChannelState(channelID, pts)
		e.chanMux.Unlock()
	}); err != nil {
		return err
	}

	e.initCommonBoxes(state)
	e.chanMux.Lock()
	for _, state := range e.channels {
		state.pts.run()
		state.recoverGap <- struct{}{}
	}
	e.chanMux.Unlock()

	defer func() {
		// Stop recover workers.
		close(e.workers)
		e.wg.Wait()

		// Stop sequence box workers.
		e.seq.stop()
		e.pts.stop()
		e.qts.stop()
		e.chanMux.Lock()
		for _, state := range e.channels {
			state.pts.stop()
		}
		e.chanMux.Unlock()

		e.cancel()
	}()

	if !e.forget {
		e.recoverGap <- struct{}{}
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for u := range e.uchan {
			if err := e.handleUpdates(u); err != nil {
				return err
			}
		}

		return nil
	})

	g.Go(func() error {
		defer func() {
			e.closed.Store(true)
			close(e.uchan)
		}()

		<-ctx.Done()
		return ctx.Err()
	})

	return g.Wait()
}

func (e *Engine) getState(ctx context.Context) (State, error) {
	if e.forget {
		if err := e.storage.ForgetAll(); err != nil {
			return State{}, err
		}

		remote, err := e.raw.UpdatesGetState(ctx)
		if err != nil {
			return State{}, xerrors.Errorf("get remote state: %w", err)
		}

		state := State{}.fromRemote(remote)
		if err := e.storage.SetState(state); err != nil {
			return State{}, xerrors.Errorf("save remote state: %w", err)
		}

		return state, nil
	}

	state, err := e.storage.GetState()
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			remote, err := e.raw.UpdatesGetState(ctx)
			if err != nil {
				return State{}, xerrors.Errorf("get remote state: %w", err)
			}

			state = state.fromRemote(remote)
			if err := e.storage.SetState(state); err != nil {
				return State{}, xerrors.Errorf("save remote state: %w", err)
			}

			return state, nil
		}

		return State{}, xerrors.Errorf("restore local state: %w", err)
	}

	return state, nil
}
