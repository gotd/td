package crypto

import (
	"crypto/rand"
	mathrand "math/rand"
	"testing"
)

func TestNewSessionID(t *testing.T) {
	t.Run("Crypto", func(t *testing.T) {
		if _, err := NewSessionID(rand.Reader); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Math", func(t *testing.T) {
		rnd := mathrand.New(mathrand.NewSource(239))
		n, err := NewSessionID(rnd)
		if err != nil {
			t.Fatal(err)
		}
		if n != 2092333242585417206 {
			t.Fatal("mismatch")
		}
	})
}
