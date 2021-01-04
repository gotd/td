package telegram

import (
	"testing"
)

func TestCondOnce(t *testing.T) {
	const n = 5

	c := createCondOnce()
	first := make(chan struct{}, 1)
	for i := 0; i < n; i++ {
		go func() {
			c.WaitIfNeeded()
			first <- struct{}{}
		}()
	}

	<-first
	select {
	case <-first:
		t.Fatal("unexpected read")
	default:
	}
	c.Done()
	for i := 0; i < n-1; i++ {
		<-first
	}
}
