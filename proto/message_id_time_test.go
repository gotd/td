package proto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"
)

// The lower 32 bits are a fraction of the second in units of 2^-32 s, so
// Time() decodes server ids (any fraction) and own ids within 1 ns.
func TestMessageIDTimeRoundTrip(t *testing.T) {
	a := require.New(t)
	sec := int64(1700000000)
	for _, ns := range []int64{0, 1, 400_000_000, 500_000_000, 900_000_000, 999_999_999} {
		now := time.Unix(sec, ns)
		got := NewMessageID(now, MessageFromServer).Time()
		a.InDelta(0, got.Sub(now), float64(4*time.Nanosecond), "%d ns", ns)
	}
	// Server id with fraction 0.9 (above 2^31): decoded wrongly before.
	id := MessageID(sec<<32 | int64(uint32(3865470566)&^3) | yieldFromServer)
	a.InDelta(0, id.Time().Sub(time.Unix(sec, 900_000_000)), float64(time.Microsecond))
}

// The lower 32 bits of generated ids span the whole range, not only
// [0, 1e9) as raw nanoseconds did (a passive fingerprint).
func TestMessageIDGenLowBitsSpread(t *testing.T) {
	clock := neo.NewTime(time.Unix(1700000000, 0))
	gen := NewMessageIDGen(clock.Now)
	var maxLow uint32
	for i := 0; i < 1000; i++ {
		clock.Travel(time.Millisecond)
		if low := uint32(gen.New(MessageFromClient)); low > maxLow {
			maxLow = low
		}
	}
	require.Greater(t, maxLow, uint32(0xe0000000), "lower bits must cover more than 90%% of the range")
}

// Ids increase even when the clock goes back (e.g. a server time offset
// correction) and keep their type bits.
func TestMessageIDGenMonotonic(t *testing.T) {
	a := require.New(t)
	clock := neo.NewTime(time.Unix(1700000000, 0))
	gen := NewMessageIDGen(clock.Now)
	var last int64
	for i := 0; i < 2000; i++ {
		switch i {
		case 500:
			clock.Travel(-10 * time.Minute)
		case 1000:
			clock.Travel(10 * time.Second)
		}
		typ := MessageFromClient
		if i%3 == 0 {
			typ = MessageFromServer
		}
		id := gen.New(typ)
		a.Greater(id, last, "step %d", i)
		a.Equal(typ, MessageID(id).Type(), "step %d", i)
		last = id
	}
}
