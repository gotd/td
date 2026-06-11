package uploader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploader_checkPartSize(t *testing.T) {
	tests := []struct {
		name     string
		partSize int
		err      bool
	}{
		{"Zero", 0, true},
		{"Not divisible by 1024", 1023, true},
		{"Max not divisible by part", MaximumPartSize + 1024, true},
		{"Default", defaultPartSize, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPartSize(tt.partSize)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_computePartSize(t *testing.T) {
	const mb = 1024 * 1024

	tests := []struct {
		name  string
		total int64
		want  int
	}{
		{"Empty", 0, defaultPartSize},
		{"Small", 1 * mb, defaultPartSize},
		// Exactly at default capacity (128 KB * 3999 ≈ 499.9 MB): stays default.
		{"FitsDefaultExact", int64(defaultPartSize) * partsLimit, defaultPartSize},
		// One byte over default capacity: must grow to the next part size.
		{"GrowsJustAbove", int64(defaultPartSize)*partsLimit + 1, 256 * 1024},
		// ~1.5 GB requires 512 KB parts.
		{"GrowsTo512K", 1536 * mb, MaximumPartSize},
		// 2 GB (Telegram max) fits with the maximum part size.
		{"MaxFile", 2 * 1024 * mb, MaximumPartSize},
		// Beyond any valid size: clamped to MaximumPartSize.
		{"Huge", 8 * 1024 * mb, MaximumPartSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computePartSize(tt.total)
			require.Equal(t, tt.want, got)
			// Computed part size must always be valid.
			require.NoError(t, checkPartSize(got))
			// And, when possible, keep parts within the limit.
			if tt.total <= int64(MaximumPartSize)*partsLimit {
				require.LessOrEqual(t, computeParts(got, int(tt.total)), partsLimit)
			}
		})
	}
}

func Test_computeParts(t *testing.T) {
	tests := []struct {
		name     string
		partSize int
		total    int
		want     int
	}{
		{"Exact part", 1024, 1024, 1},
		{"Bit more than part", 1024, 1024 + 1, 2},
		{"Stream", 1024, -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, computeParts(tt.partSize, tt.total))
		})
	}
}
