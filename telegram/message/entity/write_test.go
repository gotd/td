package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

func TestComputeLength(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{string([]rune{97, 127987, 65039, 8205, 127752}), 7},
		{string([]int32{97, 127987, 65039, 8205, 127752, 127987, 65039, 8205, 127752}), 13},
		{string([]int32{97, 128104, 8205, 128102, 8205, 128102}), 9},
	}
	for _, tt := range tests {
		testutil.ZeroAlloc(t, func() {
			_ = ComputeLength(tt.s)
		})
		t.Run(tt.s, func(t *testing.T) {
			require.Equal(t, tt.want, ComputeLength(tt.s))
		})
	}
}
