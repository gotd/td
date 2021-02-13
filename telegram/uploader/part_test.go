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
		{"524288 not divisible by part", 524288 + 1024, true},
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
