package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestBuilder_TextRange(t *testing.T) {
	var (
		a = require.New(t)
		b Builder
	)
	_, _ = b.WriteString("abc")
	a.Equal("abc"[1:2], b.TextRange(1, 2))
	a.Equal("abc"[0:0], b.TextRange(0, 0))

	panicRanges := [][2]int{
		{1, 0},
		{-1, 0},
		{0, -1},
	}
	for _, r := range panicRanges {
		a.Panics(func() {
			b.TextRange(r[0], r[1])
		})
	}
}

func TestBuilder_LastEntity(t *testing.T) {
	var (
		a = require.New(t)
		b Builder
	)

	e, ok := b.LastEntity()
	a.False(ok)
	a.Nil(e)
	b.Underline("abc")
	e, ok = b.LastEntity()
	a.True(ok)
	a.Equal(&tg.MessageEntityUnderline{
		Offset: 0,
		Length: 3,
	}, e)
}

func TestBuilder_GrowText(t *testing.T) {
	var (
		a = require.New(t)
		b Builder
	)

	b.GrowText(100)
	a.LessOrEqual(100, b.message.Cap())
}

func TestBuilder_GrowEntities(t *testing.T) {
	var (
		a = require.New(t)
		b Builder
	)

	b.GrowEntities(100)
	a.Equal(100, cap(b.entities))
	a.Panics(func() {
		b.GrowEntities(-1)
	})
}
