//go:build arm && cgo && (linux || android)

package crypto

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/ige"
)

func TestHWIGEDecryptMatchesSoftware(t *testing.T) {
	key, iv, src := newHWIGETestVector(16 * 16)

	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	want := make([]byte, len(src))
	ige.DecryptBlocks(block, iv, want, src)

	got := make([]byte, len(src))
	if !hwIGEDecrypt(key, iv, got, src) {
		t.Skip("ARMv8 AES extension is not available")
	}
	require.Equal(t, want, got)
}

func BenchmarkHWIGEDecrypt(b *testing.B) {
	key, iv, src := newHWIGETestVector(1024 * 1024)
	dst := make([]byte, len(src))
	if !hwIGEDecrypt(key, iv, dst, src) {
		b.Skip("ARMv8 AES extension is not available")
	}

	b.SetBytes(int64(len(src)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hwIGEDecrypt(key, iv, dst, src)
	}
}

func newHWIGETestVector(size int) ([]byte, []byte, []byte) {
	key := make([]byte, 32)
	iv := make([]byte, 32)
	src := make([]byte, size)
	for i := range key {
		key[i] = byte(i*3 + 1)
	}
	for i := range iv {
		iv[i] = byte(i*5 + 2)
	}
	for i := range src {
		src[i] = byte(i*7 + 3)
	}
	return key, iv, src
}
