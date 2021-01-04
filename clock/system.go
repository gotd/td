package clock

import "time"

type systemClock struct{}

func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (systemClock) Now() time.Time                         { return time.Now() }

// System Clock.
var System Clock = systemClock{} // nolint:gochecknoglobals
