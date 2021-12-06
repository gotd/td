package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestEnsureTrim(t *testing.T) {
	a := require.New(t)

	prefix := "pre"
	expected := "abc\nabc"
	b := Builder{}
	b.Plain(prefix)
	b.Format(expected+"\n\n\n", Bold(), Italic())

	msg, ent := b.Complete()
	a.Equal(prefix+expected, msg)
	a.Equal([]tg.MessageEntityClass{
		&tg.MessageEntityBold{
			Offset: len(prefix),
			Length: ComputeLength(expected),
		},
		&tg.MessageEntityItalic{
			Offset: len(prefix),
			Length: ComputeLength(expected),
		},
	}, ent)
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
				Offset: ComputeLength("boldplain"),
				Length: ComputeLength("bold2"),
			},
			&tg.MessageEntityBold{
				Offset: 0,
				Length: ComputeLength("bold"),
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
