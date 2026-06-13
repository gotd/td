package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountPadding(t *testing.T) {
	a := require.New(t)

	for l := range 4096 {
		for _, randByte := range []byte{0x00, 0x01, 0x0F, 0x7F, 0xF0, 0xFF} {
			padding := countPadding(l, randByte)
			total := l + padding

			// MTProto 2.0 requires 12..1024 bytes of padding.
			a.GreaterOrEqual(padding, 12, "l=%d rand=%#x", l, randByte)
			a.LessOrEqual(padding, 1024, "l=%d rand=%#x", l, randByte)
			// Encrypted length must be a multiple of the 16-byte block size.
			a.Zero(total%16, "l=%d rand=%#x", l, randByte)
			// Only the low 4 bits of randByte select the random component.
			a.Equal(countPadding(l, randByte&0x0F), padding, "l=%d rand=%#x", l, randByte)
		}
	}
}

func TestCountPaddingJitter(t *testing.T) {
	a := require.New(t)

	// The random component must change the total length, otherwise the
	// encrypted-message size would be a deterministic fingerprint.
	base := countPadding(64, 0x00)
	jittered := countPadding(64, 0x0F)
	a.Equal(base+0x0F*16, jittered)
	a.NotEqual(base, jittered)
}
