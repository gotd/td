package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFillBytes(t *testing.T) {
	tests := []struct {
		name    string
		bitSize int
		result  bool
	}{
		{"Smaller", 255, true},
		{"Equal", 256, true},
		{"Bigger", 512, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			b, err := rand.Prime(rand.Reader, tt.bitSize)
			a.NoError(err)

			var to [256 / 8]byte
			a.Equal(tt.result, FillBytes(b, to[:]))
		})
	}
}
