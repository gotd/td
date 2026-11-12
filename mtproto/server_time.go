package mtproto

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/clock"
)

// ServerTime tracks the difference between the server clock and the local
// clock, so that message ids and the msg_id time window use server time even
// when the local clock is off.
//
// Like TDLib (AuthData::update_server_time_difference), the difference is
// taken from the time encoded in every received msg_id (and the server_time
// of the key exchange): the first observation sets it, later ones only raise
// it, since a message is always received after the server created it. A
// bad_msg_notification "msg_id too high" resets it from the notification's
// own msg_id.
//
// One ServerTime is shared by all connections of a client. It is safe for
// concurrent use.
type ServerTime struct {
	base clock.Clock

	mux sync.Mutex
	// offset is server time minus local time, in nanoseconds.
	offset atomic.Int64
	// known reports whether offset was observed (not only seeded).
	known bool
}

// NewServerTime creates a ServerTime over the base clock (clock.System if
// nil), starting with the given offset, e.g. a persisted one. The offset is
// replaced by the first observed server time.
func NewServerTime(base clock.Clock, offset time.Duration) *ServerTime {
	if base == nil {
		base = clock.System
	}
	t := &ServerTime{base: base}
	t.offset.Store(int64(offset))
	return t
}

// Now returns the current server time.
func (t *ServerTime) Now() time.Time {
	return t.base.Now().Add(time.Duration(t.offset.Load()))
}

// Offset returns server time minus local time.
func (t *ServerTime) Offset() time.Duration {
	return time.Duration(t.offset.Load())
}

// Seed sets the offset, e.g. one persisted by a previous run, unless it was
// already observed from the server.
func (t *ServerTime) Seed(offset time.Duration) {
	if t.Known() {
		return
	}
	t.offset.Store(int64(offset))
}

// Known reports whether the offset was observed from the server.
func (t *ServerTime) Known() bool {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.known
}

// minOffsetStep is the smallest increase of the offset that is applied, as in
// TDLib (1e-4 s).
const minOffsetStep = 100 * time.Microsecond

// observe updates the offset from a server time point received now. It
// reports whether the offset changed.
func (t *ServerTime) observe(server time.Time) bool {
	diff := int64(server.Sub(t.base.Now()))

	t.mux.Lock()
	defer t.mux.Unlock()
	if !t.known {
		t.known = true
		t.offset.Store(diff)
		return true
	}
	if diff > t.offset.Load()+int64(minOffsetStep) {
		t.offset.Store(diff)
		return true
	}
	return false
}

// reset sets the offset from a server time point received now, also when it
// is lower than the current one (the local clock went forward or the offset
// was too high).
func (t *ServerTime) reset(server time.Time) {
	diff := int64(server.Sub(t.base.Now()))

	t.mux.Lock()
	defer t.mux.Unlock()
	t.known = true
	t.offset.Store(diff)
}
