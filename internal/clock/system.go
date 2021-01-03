package clock

import "time"

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

// System Clock.
var System Clock = systemClock{} // nolint:gochecknoglobals
