package proto

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/testutil"
)

// Replays of any of the last N ids are rejected, not only of the latest one
// (the old buffer effectively kept a single id).
func TestMessageIDBufReplay(t *testing.T) {
	a := require.New(t)
	b := NewMessageIDBuf(100)
	for _, id := range []int64{1000, 1004, 1008, 1012} {
		a.True(b.Consume(id), "first consume of %d", id)
	}
	a.False(b.Consume(1004), "replay of a stored id")
	a.False(b.Consume(1000), "replay of the oldest stored id")
	a.True(b.Consume(1002), "unseen id between stored ones")
	a.False(b.Consume(1002), "replay of the inserted id")
}

func TestMessageIDBufReorder(t *testing.T) {
	a := require.New(t)
	b := NewMessageIDBuf(4)
	// Out of order arrival within the window is accepted once.
	for _, id := range []int64{20, 10, 30, 15, 25} {
		a.True(b.Consume(id), "%d", id)
	}
	for _, id := range []int64{20, 10, 30, 15, 25} {
		a.False(b.Consume(id), "replay %d", id)
	}
}

// Once N ids are stored, an id lower than all of the last N is rejected.
func TestMessageIDBufWindow(t *testing.T) {
	a := require.New(t)
	const n = 10
	b := NewMessageIDBuf(n)
	for i := int64(1); i <= 5*n; i++ {
		a.True(b.Consume(i * 4))
	}
	// The newest N ids (164..200) are always kept.
	for i := int64(4*n + 1); i <= 5*n; i++ {
		a.False(b.Consume(i*4), "recent replay %d", i*4)
	}
	a.False(b.Consume(4), "far below the window")
	a.False(b.Consume(-1))
	a.False(b.Consume(0))

	b.Reset()
	a.True(b.Consume(4), "accepted after Reset")
}

func TestMessageIDBufZeroAlloc(t *testing.T) {
	b := NewMessageIDBuf(100)
	var id int64 = 1 << 40
	testutil.ZeroAlloc(t, func() {
		id += 4
		b.Consume(id)
		b.Consume(id - 2) // insert path
		b.Consume(id)     // duplicate
	})
}

func BenchmarkMessageIDBufIncreasing(b *testing.B) {
	buf := NewMessageIDBuf(100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Consume(int64(i + 1))
	}
}
