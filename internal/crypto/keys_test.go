package crypto

import (
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/testutil"
)

func BenchmarkKeys(b *testing.B) {
	var k Key
	for i := 0; i < 256; i++ {
		k[i] = byte(i)
	}
	var m bin.Int128
	for i := 0; i < 16; i++ {
		m[i] = byte(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Keys(k, m, Client)
	}
}

func TestKeys(t *testing.T) {
	var k Key
	for i := 0; i < 256; i++ {
		k[i] = byte(i)
	}
	var m bin.Int128
	for i := 0; i < 16; i++ {
		m[i] = byte(i)
	}

	testutil.ZeroAlloc(t, func() {
		_, _ = Keys(k, m, Client)
	})
}
