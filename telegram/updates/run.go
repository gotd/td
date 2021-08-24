package updates

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// HandleUpdates handles updates.
func (e *Engine) HandleUpdates(u tg.UpdatesClass) error {
	e.shutdownMux.Lock()
	closed := e.closed
	e.shutdownMux.Unlock()
	if closed {
		return xerrors.New("closed")
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

	e.initCommonBoxes(state)
	if err := func() error {
		e.chanMux.Lock()
		defer e.chanMux.Unlock()

		return e.storage.Channels(func(channelID, pts int) {
			e.channels[channelID] = e.createChannelState(channelID, pts)
		})
	}(); err != nil {
		return err
	}

	defer func() {
		e.stopCommonBoxes()

		e.chanMux.Lock()
		for _, state := range e.channels {
			state.stop()
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
			e.shutdownMux.Lock()
			close(e.uchan)
			e.closed = true
			e.shutdownMux.Unlock()
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
		if xerrors.Is(err, ErrStateNotFound) {
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
