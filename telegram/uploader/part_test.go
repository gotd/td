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
		// Exactly at max capacity (512 KB * 3999 ≈ 1.95 GB): largest size that
		// still fits within the parts limit at the maximum part size.
		{"MaxFile", int64(MaximumPartSize) * partsLimit, MaximumPartSize},
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
				require.LessOrEqual(t, computeParts(got, tt.total), partsLimit)
			}
		})
	}
}

func Test_computeParts(t *testing.T) {
	tests := []struct {
		name     string
		partSize int
		total    int64
		want     int
	}{
		{"Exact part", 1024, 1024, 1},
		{"Bit more than part", 1024, 1024 + 1, 2},
		{"Stream", 1024, -1, 0},
		// Total exceeding int32 must not overflow the part-count math.
		{"Over32Bit", MaximumPartSize, 8 * 1024 * 1024 * 1024, 16384},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, computeParts(tt.partSize, tt.total))
		})
	}
}
