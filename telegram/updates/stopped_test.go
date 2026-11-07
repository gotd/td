package updates

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

// TestManagerStopped checks that calls after Run stopped do not block
// forever, even with a context that is never canceled.
func TestManagerStopped(t *testing.T) {
	tt := newTrackTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- tt.m.Run(ctx, tt.api, trackUserID, AuthOptions{
			OnStart: func(context.Context) { close(started) },
		})
	}()
	<-started
	cancel()
	<-done

	// Never canceled: calls must return on their own.
	bg := context.Background()
	returned := make(chan error, 1)
	go func() {
		// More updates than the queue holds.
		for i := 0; i < 100; i++ {
			if err := tt.m.Handle(bg, &tg.Updates{Date: 1}); err != nil {
				returned <- err
				return
			}
		}
		returned <- nil
	}()
	select {
	case err := <-returned:
		require.ErrorIs(t, err, ErrStopped)
	case <-time.After(5 * time.Second):
		t.Fatal("Handle blocks after Run stopped")
	}

	go func() {
		for i := 0; i < 100; i++ {
			if err := tt.m.HandleAffected(bg, 0, 10+i, 1); err != nil {
				returned <- err
				return
			}
		}
		returned <- nil
	}()
	select {
	case err := <-returned:
		require.ErrorIs(t, err, ErrStopped)
	case <-time.After(5 * time.Second):
		t.Fatal("HandleAffected blocks after Run stopped")
	}

	// Manager.Run returned: the request would never be applied, as a new Run
	// requires Reset.
	go func() { returned <- tt.m.TrackChannel(bg, 42, 4242, 100) }()
	select {
	case err := <-returned:
		require.ErrorIs(t, err, ErrStopped)
	case <-time.After(5 * time.Second):
		t.Fatal("TrackChannel blocks after Run stopped")
	}
}

// TestManagerStoppedTrackRace checks the window between internalState.Run
// return and Manager.Run return, when the manager still counts as started.
func TestManagerStoppedTrackRace(t *testing.T) {
	tt := newTrackTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- tt.m.Run(ctx, tt.api, trackUserID, AuthOptions{
			OnStart: func(context.Context) { close(started) },
		})
	}()
	<-started
	cancel()
	<-done

	tt.m.mux.Lock()
	tt.m.started = true
	tt.m.mux.Unlock()
	returned := make(chan error, 1)
	go func() { returned <- tt.m.TrackChannel(context.Background(), 42, 4242, 100) }()
	select {
	case err := <-returned:
		require.ErrorIs(t, err, ErrStopped)
	case <-time.After(5 * time.Second):
		t.Fatal("TrackChannel blocks after Run stopped")
	}
}
