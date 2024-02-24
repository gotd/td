package entity

import (
	"encoding/hex"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/testutil"
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
		r := []byte(tt.s)
		testutil.ZeroAlloc(t, func() {
			_ = ComputeLength(tt.s)
		})
		testutil.ZeroAlloc(t, func() {
			_ = ComputeLengthBytes(r)
		})
		t.Run(hex.EncodeToString([]byte(tt.s)), func(t *testing.T) {
			require.Equal(t, tt.want, ComputeLength(tt.s))
			require.Equal(t, tt.want, ComputeLengthBytes(r))
		})
	}
}

func TestBuilder_Write(t *testing.T) {
	var (
		a = require.New(t)
		b Builder
	)
	_, err := b.Write([]byte("abc"))
	a.NoError(err)
	_, err = b.WriteString("abc")
	a.NoError(err)
	a.NoError(b.WriteByte('\n'))
	a.Equal(3+3+1, b.UTF8Len())
	a.Equal(3+3+1, b.UTF16Len())

	var r rune = 127987
	_, err = b.WriteRune(r)
	a.NoError(err)
	a.Equal(3+3+1+utf8.RuneLen(r), b.UTF8Len())
	a.Equal(3+3+1+utf16RuneLen(r), b.UTF16Len())
}
