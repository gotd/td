package telegram

import (
	"testing"
)

func TestCondOnce(t *testing.T) {
	const n = 5

	c := createCondOnce()
	first := make(chan struct{}, 1)
	for i := range [n]struct{}{} {
		go func(n int) {
			c.WaitIfNeeded()
			first <- struct{}{}
		}(i)
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
