package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/tg"
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

func TestComplete(t *testing.T) {
	tests := []struct {
		name     string
		format   func(e *Builder)
		msg      string
		entities []tg.MessageEntityClass
	}{
		{"PlainBold", func(e *Builder) {
			e.Plain("plain").Bold("bold")
		}, "plainbold", []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Offset: ComputeLength("plain"),
				Length: ComputeLength("bold"),
			},
		}},
		{"PlainBoldAndStrike", func(e *Builder) {
			e.Plain("plain").Format("10\n\n\n\n", Bold(), Strike())
		}, "plain10", []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Offset: ComputeLength("plain"),
				Length: ComputeLength("10"),
			},
			&tg.MessageEntityStrike{
				Offset: ComputeLength("plain"),
				Length: ComputeLength("10"),
			},
		}},
		{"BoldPlainBold", func(e *Builder) {
			e.Bold("bold").Plain("plain").Bold("bold2\n\n\n\n")
		}, "boldplainbold2", []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Offset: 0,
				Length: ComputeLength("bold"),
			},
			&tg.MessageEntityBold{
				Offset: ComputeLength("boldplain"),
				Length: ComputeLength("bold2"),
			},
		}},
		{"BoldBold", func(e *Builder) {
			e.Bold("bold\n\n\n\n").Bold("bold2\n\n\n\n")
		}, "bold\n\n\n\nbold2", []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Offset: 0,
				Length: ComputeLength("bold\n\n\n\n"),
			},
			&tg.MessageEntityBold{
				Offset: ComputeLength("bold\n\n\n\n"),
				Length: ComputeLength("bold2"),
			},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)
			b := Builder{}
			test.format(&b)

			msg, entities := b.Complete()
			a.Equal(test.msg, msg)
			a.Equal(test.entities, entities)
		})
	}
}
