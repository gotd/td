package updates

import (
	"context"
	"errors"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// HandleUpdates handles updates.
func (e *Engine) HandleUpdates(u tg.UpdatesClass) error {
	if e.closed.Load() {
		return xerrors.Errorf("closed")
	}
	e.wg.Add(1)
	defer e.wg.Done()
	return e.handleUpdates(u)
}

// Run starts the engine and calls f after initialization.
func (e *Engine) Run(ctx context.Context, f func(context.Context) error) error {
	if e.closed.Load() {
		return xerrors.Errorf("closed")
	}

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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		e.closed.Store(true)
		close(e.workers)
		e.wg.Wait()

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

	e.initCommonBoxes(state)

	e.chanMux.Lock()
	for _, state := range e.channels {
		state.pts.run()
		state.recoverGap <- struct{}{}
	}
	e.chanMux.Unlock()

	if !e.forget {
		if err := e.recoverState(); err != nil {
			return xerrors.Errorf("recover common state: %w", err)
		}
	}

	go func() { e.echan <- f(ctx) }()
	return <-e.echan
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
