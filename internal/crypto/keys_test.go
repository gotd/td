package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/testutil"
)

func genMessageAndAuthKeys() (Key, bin.Int128) {
	var k Key
	for i := 0; i < 256; i++ {
		k[i] = byte(i)
	}
	var m bin.Int128
	for i := 0; i < 16; i++ {
		m[i] = byte(i)
	}

	return k, m
}

func BenchmarkKeys(b *testing.B) {
	k, m := genMessageAndAuthKeys()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Keys(k, m, Client)
	}
}

func TestKeys(t *testing.T) {
	k, m := genMessageAndAuthKeys()

	testutil.ZeroAlloc(t, func() {
		_, _ = Keys(k, m, Client)
	})
}

func BenchmarkMessageKey(b *testing.B) {
	k, _ := genMessageAndAuthKeys()
	payload := make([]byte, 1024)
	if _, err := io.ReadFull(rand.Reader, payload); err != nil {
		b.Error(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = MessageKey(k, payload, Client)
	}
}

func TestMessageKey(t *testing.T) {
	k, _ := genMessageAndAuthKeys()
	payload := make([]byte, 1024)
	if _, err := io.ReadFull(rand.Reader, payload); err != nil {
		t.Error(err)
	}

	testutil.ZeroAlloc(t, func() {
		_ = MessageKey(k, payload, Client)
	})
}
