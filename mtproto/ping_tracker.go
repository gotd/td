package mtproto

import "time"

// pingTracker decides whether an outstanding ping means the connection is dead.
//
// It is a pure state machine: the ping loop samples byte and message counters
// with its own clock and feeds the values in. Any change is liveness at the
// sample time, so there is a single time source and tests stay deterministic.
type pingTracker struct {
	timeout, maxWait time.Duration
	pendingSince     time.Time // zero: no unanswered ping
	lastAlive        time.Time
	lastProgress     uint64
	lastPongSeq      uint64
}

func newPingTracker(now time.Time, timeout, maxWait time.Duration, progress, pongSeq uint64) *pingTracker {
	return &pingTracker{
		timeout:      timeout,
		maxWait:      maxWait,
		lastAlive:    now,
		lastProgress: progress,
		lastPongSeq:  pongSeq,
	}
}

// observe samples activity counters; any change counts as liveness at now.
func (t *pingTracker) observe(now time.Time, progress, pongSeq uint64) {
	if progress != t.lastProgress {
		t.lastProgress = progress
		t.lastAlive = now
	}
	if pongSeq != t.lastPongSeq {
		t.lastPongSeq = pongSeq
		t.lastAlive = now
		t.pendingSince = time.Time{}
	}
}

func (t *pingTracker) onSend(now time.Time) {
	if t.pendingSince.IsZero() {
		t.pendingSince = now
	}
}

func (t *pingTracker) dead(now time.Time) (reason string, dead bool) {
	if t.pendingSince.IsZero() {
		return "", false
	}
	ref := t.pendingSince
	if t.lastAlive.After(ref) {
		ref = t.lastAlive
	}
	if now.Sub(ref) > t.timeout {
		return "no_progress", true
	}
	if t.maxWait > 0 && now.Sub(t.pendingSince) > t.maxWait {
		return "max_wait", true
	}
	return "", false
}

// pingCheckInterval is how often progress-aware mode samples liveness.
// The worst-case detection delay is PingTimeout plus this interval.
func pingCheckInterval(timeout time.Duration) time.Duration {
	d := timeout / 5
	if d < 100*time.Millisecond {
		d = 100 * time.Millisecond
	}
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
