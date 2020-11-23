package crypto

import (
	"crypto/sha256"
	"testing"
)

func BenchmarkSHA256A(b *testing.B) {
	var (
		authKey = make([]byte, 256)
		msgKey  = make([]byte, 16)
		buf     = make([]byte, 0, sha256.Size)
	)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = SHA256A(buf, authKey, msgKey, ModeServer)
		buf = buf[:0]
	}
}
