package tdesktop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_tdesktopMD5(t *testing.T) {
	tests := []struct {
		input, output string
	}{
		{"data", "D877F783D5D3EF8C18D5027F940662CD"},
		{"data#1", "438B3BB129B86F4E9CA19D4CDF58C398"},
		{"data#2", "A7FDF864FBC10B7772ADA161AE3C900E"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.output, tdesktopMD5(tt.input))
		})
	}
}
