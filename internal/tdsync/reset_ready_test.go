package tdsync

import (
	"sync"
	"testing"
)

func TestResetReady(t *testing.T) {
	t.Run("Ready", func(t *testing.T) {
		r := NewResetReady()

		wait := make(chan struct{})
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-r.Ready()
			<-r.Ready()
			close(wait)
		}()

		// Check that Ready can be called multiple times
		// from different threads.
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-r.Ready()
			<-r.Ready()
		}()

		// Check that Signal can be called multiple times
		// from different threads.
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-wait
			r.Signal()
		}()

		// Check Signal call logic.
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Signal()
			r.Signal()
			<-wait
		}()

		wg.Wait()
	})

	checkNoSignal := func(t *testing.T, r *ResetReady) {
		select {
		case <-r.Ready():
			t.Error("unexpected signal")
		default:
		}
	}

	t.Run("Reset", func(t *testing.T) {
		t.Run("Zero", func(t *testing.T) {
			r := NewResetReady()
			checkNoSignal(t, r)

			acquire := make(chan struct{})
			release := make(chan struct{})
			go func() {
				close(acquire)
				<-r.Ready()
				close(release)
			}()

			<-acquire
			r.Reset()
			<-release
			checkNoSignal(t, r)
		})

		t.Run("NoSignal", func(t *testing.T) {
			r := NewResetReady()
			checkNoSignal(t, r)
			r.Reset()
			checkNoSignal(t, r)
		})

		t.Run("Signal", func(t *testing.T) {
			r := NewResetReady()

			wait := make(chan struct{})
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-r.Ready()
				close(wait)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.Signal()
				<-wait
			}()

			wg.Wait()
			r.Reset()
			checkNoSignal(t, r)
		})
	})
}
