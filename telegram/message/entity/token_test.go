package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestToken_Apply(t *testing.T) {
	var (
		a = require.New(t)
		b = &Builder{}
	)
	_ = b.WriteByte('a')
	tok := b.Token()
	a.Equal(1, tok.UTF8Offset())
	a.Equal(1, tok.UTF16Offset())

	a.Zero(tok.UTF8Length(b))
	a.Zero(tok.UTF16Length(b))
	a.Empty(tok.Text(b))

	text := "abc🏳"
	_, _ = b.WriteString(text)
	a.Equal(text, tok.Text(b))
	utf16Len := ComputeLength(tok.Text(b))

	a.Equal(b.message.Len()-tok.UTF8Offset(), tok.UTF8Length(b))
	a.Equal(utf16Len, tok.UTF16Length(b))

	tok.Apply(b, Bold())
	a.Equal(1, b.EntitiesLen())
	e, ok := b.LastEntity()
	a.True(ok)
	a.Equal(&tg.MessageEntityBold{Offset: 1, Length: utf16Len}, e)
}
