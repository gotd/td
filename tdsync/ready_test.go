package tdsync

import (
	"sync"
	"testing"
)

func TestReady(t *testing.T) {
	r := NewReady()

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
}
