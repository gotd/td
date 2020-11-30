package proto

import (
	"testing"
	"time"
)

func TestMessageID(t *testing.T) {
	now := time.Date(2018, 10, 10, 23, 42, 6, 13600, time.UTC)
	id := MessageID(newMessageID(now, 0))
	if id.Type() != MessageFromClient {
		t.Fatal("invalid type")
	}
	if id != 6610877768685073696 {
		t.Error("mismatch")
	}
	delta := id.Time().Sub(now)
	if delta < 0 {
		delta *= -1
	}
	if delta > time.Second {
		t.Fatal("unexpected time drift")
	}
	t.Run("NewMessageID", func(t *testing.T) {
		if NewMessageID(now, MessageFromServer).Type() != MessageFromServer {
			t.Error("Mismatch")
		}
		if NewMessageID(now, 100).Type() != MessageFromClient {
			t.Error("Mismatch")
		}
	})
}

func BenchmarkNewMessageID(b *testing.B) {
	// Note that most overhead will be from time.Now() calls.
	// Just ensuring that NewMessageID itself is reasonably fast.
	now := time.Date(2018, 10, 10, 23, 42, 6, 13600, time.UTC)
	for i := 0; i < b.N; i++ {
		if NewMessageID(now, MessageFromServer).Type() != MessageFromServer {
			b.Fatal("Mismatch")
		}
	}
}
