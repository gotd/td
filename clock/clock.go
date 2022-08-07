// Package clock abstracts time source.
package clock

import (
	"time"

	"github.com/gotd/neo"
)

// Clock is current time source.
type Clock interface {
	Now() time.Time
	Timer(d time.Duration) Timer
	Ticker(d time.Duration) Ticker
}

// Timer abstracts a single event.
type Timer = neo.Timer

// StopTimer stops timer and drains timer channel if Stop() returned false.
func StopTimer(t Timer) {
	if !t.Stop() {
		select {
		case <-t.C():
		default:
		}
	}
}

// Ticker abstracts a channel that delivers “ticks” of a clock at intervals.
type Ticker = neo.Ticker
