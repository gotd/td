package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPingTrackerIdle(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	tr := newPingTracker(t0, 15*time.Second, 0, 0, 0)

	_, dead := tr.dead(t0.Add(10 * time.Minute))
	require.False(t, dead)
}

func TestPingTrackerNoProgress(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	tr := newPingTracker(t0, timeout, 0, 0, 0)
	tr.onSend(t0)

	_, dead := tr.dead(t0.Add(timeout))
	require.False(t, dead, "exactly PingTimeout is still alive")

	reason, dead := tr.dead(t0.Add(timeout + time.Millisecond))
	require.True(t, dead)
	require.Equal(t, "no_progress", reason)
}

func TestPingTrackerProgressExtendsThenExpires(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	tr := newPingTracker(t0, timeout, 0, 0, 0)
	tr.onSend(t0)

	now := t0
	progress := tr.lastProgress
	for i := 0; i < 20; i++ { // 20 * timeout/2 == 10 * timeout
		now = now.Add(timeout / 2)
		progress++
		tr.observe(now, progress, tr.lastPongSeq)
		_, dead := tr.dead(now)
		require.False(t, dead)
	}

	reason, dead := tr.dead(now.Add(timeout + time.Millisecond))
	require.True(t, dead)
	require.Equal(t, "no_progress", reason)
}

func TestPingTrackerProgressBeforeSendDoesNotExtend(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	// Progress is sampled at t0-1s, then the ping is sent at t0.
	// observe runs before onSend, so that sample must not push the deadline
	// past the send time: dead uses max(pendingSince, lastAlive).
	tr := newPingTracker(t0.Add(-time.Second), timeout, 0, 0, 0)
	tr.observe(t0.Add(-time.Second), 1, 0)
	tr.onSend(t0)

	reason, dead := tr.dead(t0.Add(timeout + time.Millisecond))
	require.True(t, dead)
	require.Equal(t, "no_progress", reason)
}

func TestPingTrackerAnyPongClearsPending(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	tr := newPingTracker(t0, timeout, 0, 0, 0)
	tr.onSend(t0)
	tr.observe(t0.Add(time.Millisecond), tr.lastProgress, tr.lastPongSeq+1)

	_, dead := tr.dead(t0.Add(10 * timeout))
	require.False(t, dead)
	require.True(t, tr.pendingSince.IsZero())
}

func TestPingTrackerMultiplePingsKeepEarliest(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	interval := time.Minute
	tr := newPingTracker(t0, 15*time.Second, 0, 0, 0)
	tr.onSend(t0)
	tr.onSend(t0.Add(interval))

	require.True(t, tr.pendingSince.Equal(t0))
}

func TestPingTrackerMaxWait(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	maxWait := 3 * timeout
	tr := newPingTracker(t0, timeout, maxWait, 0, 0)
	tr.onSend(t0)

	now := t0
	progress := uint64(0)
	for now.Sub(t0) < maxWait {
		now = now.Add(timeout / 2)
		progress++
		tr.observe(now, progress, 0)
		_, dead := tr.dead(now)
		require.False(t, dead)
	}

	_, dead := tr.dead(t0.Add(maxWait))
	require.False(t, dead, "exactly PingMaxWait is still alive")

	reason, dead := tr.dead(t0.Add(maxWait + time.Millisecond))
	require.True(t, dead)
	require.Equal(t, "max_wait", reason)
}

func TestPingTrackerMaxWaitDisabled(t *testing.T) {
	t0 := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	timeout := 15 * time.Second
	tr := newPingTracker(t0, timeout, 0, 0, 0)
	tr.onSend(t0)

	now := t0
	progress := uint64(0)
	end := t0.Add(100 * timeout)
	for now.Before(end) {
		now = now.Add(timeout / 2)
		progress++
		tr.observe(now, progress, 0)
		_, dead := tr.dead(now)
		require.False(t, dead)
	}
}

func TestPingCheckInterval(t *testing.T) {
	require.Equal(t, 100*time.Millisecond, pingCheckInterval(100*time.Millisecond))
	require.Equal(t, 3*time.Second, pingCheckInterval(15*time.Second))
	require.Equal(t, 5*time.Second, pingCheckInterval(60*time.Second))
}
