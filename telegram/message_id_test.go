package telegram

import (
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/gotd/td/internal/proto"
)

func TestMessageIDGen(t *testing.T) {
	date := time.Date(1991, 1, 3, 14, 44, 33, 513, time.UTC)
	clock := neo.NewTime(date)

	gen := newMsgIDGen(clock.Now, 10, proto.MessageFromServer)
	met := make(map[int64]bool)

	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			clock.Travel(time.Millisecond * 100)
		}

		id := gen.New()
		if met[id] {
			t.Fatal("met")
		}

		met[id] = true
	}
}

func BenchmarkMsgIDGen_New(b *testing.B) {
	b.ReportAllocs()

	date := time.Date(1991, 1, 3, 14, 44, 33, 513, time.UTC)
	var dateCalls int
	now := func() time.Time {
		if dateCalls%100 == 0 {
			date = date.Add(time.Millisecond)
		}
		return date
	}

	gen := newMsgIDGen(now, 50, proto.MessageFromClient)

	for i := 0; i < b.N; i++ {
		_ = gen.New()
	}
}
