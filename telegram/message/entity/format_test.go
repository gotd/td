package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

func Test_computeLength(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"a🏳️‍🌈", 7},
		{"a🏳️‍🌈🏳️‍🌈", 13},
		{"a🏳️‍🌈🏳️‍🌈a🏳️‍🌈🏳️‍🌈", 26},
		{"a👨‍👦‍👦", 9},
		{`message#bce383d2
  id: 1939
  from_id: 🏳️‍🌈
  date: 2021-03-15T10:01:41Z`, 74},
	}
	for _, tt := range tests {
		testutil.ZeroAlloc(t, func() {
			computeLength(tt.s)
		})
		t.Run(tt.s, func(t *testing.T) {
			require.Equal(t, tt.want, computeLength(tt.s))
		})
	}
}
