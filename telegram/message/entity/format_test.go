package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/tg"
)

func Test_computeLength(t *testing.T) {
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

func TestEnsureTrim(t *testing.T) {
	a := require.New(t)

	prefix := "pre"
	expected := "abc\nabc"
	b := Builder{}
	b.Plain(prefix)
	b.Format(expected+"\n\n\n", Bold(), Italic())

	msg, ent := b.Complete()
	a.Equal(prefix+expected, msg)
	a.Len(ent, 2)
	a.Equal(&tg.MessageEntityBold{
		Offset: len(prefix),
		Length: ComputeLength(expected),
	}, ent[0])
	a.Equal(&tg.MessageEntityItalic{
		Offset: len(prefix),
		Length: ComputeLength(expected),
	}, ent[1])
}
