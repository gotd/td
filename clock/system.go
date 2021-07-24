package clock

import "time"

type systemTimer struct {
	timer *time.Timer
}

func (s systemTimer) C() <-chan time.Time {
	return s.timer.C
}

func (s systemTimer) Reset(d time.Duration) {
	s.timer.Reset(d)
}

func (s systemTimer) Stop() bool {
	return s.timer.Stop()
}

type systemTicker struct {
	ticker *time.Ticker
}

func (s systemTicker) C() <-chan time.Time {
	return s.ticker.C
}

func (s systemTicker) Stop() {
	s.ticker.Stop()
}

func (s systemTicker) Reset(d time.Duration) {
	s.ticker.Reset(d)
}

type systemClock struct{}

func (clock systemClock) Ticker(d time.Duration) Ticker {
	return systemTicker{ticker: time.NewTicker(d)}
}

func (systemClock) Timer(d time.Duration) Timer {
	return systemTimer{timer: time.NewTimer(d)}
}

func (systemClock) Now() time.Time { return time.Now() }

// System Clock.
var System Clock = systemClock{}
