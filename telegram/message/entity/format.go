package entity

import (
	"strings"

	"github.com/gotd/td/tg"
)

type utf8entity struct {
	offset int
	length int
}

// Builder builds message string and text entities.
type Builder struct {
	entities []tg.MessageEntityClass
	// lengths stores offset/length of entities too, but in UTF-8 codepoints.
	lengths []utf8entity
	// We store index of first entity added at last Format call.
	// It needed to trim space in all entities of last text block.
	lastFormatIndex int
	// utf16length stores length in UTF-16 codepoints.
	utf16length int
	// message is message string builder.
	message strings.Builder
}

// GrowText grows internal buffer capacity.
func (b *Builder) GrowText(n int) {
	b.message.Grow(n)
}

// GrowEntities grows internal buffer capacity.
func (b *Builder) GrowEntities(n int) {
	if n < 0 {
		panic("entity.Builder.GrowEntities: negative count")
	}

	buf := make([]tg.MessageEntityClass, len(b.entities), 2*cap(b.entities)+n)
	copy(buf, b.entities)
	b.entities = buf
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.message.Reset()
	b.entities = nil
	b.utf16length = 0
}

// UTF8Len returns length of text in bytes.
func (b *Builder) UTF8Len() int {
	return b.message.Len()
}

// UTF16Len returns length of text in UTF-16 codepoints.
func (b *Builder) UTF16Len() int {
	return b.utf16length
}

// EntitiesLen return length of added entities.
func (b *Builder) EntitiesLen() int {
	return len(b.entities)
}

// TextRange returns message text of given byte (UTF-8) range.
//
// If range is invalid, it will panic.
func (b *Builder) TextRange(from, to int) string {
	return b.message.String()[from:to]
}

// LastEntity returns last entity if any.
func (b *Builder) LastEntity() (tg.MessageEntityClass, bool) {
	l := b.EntitiesLen()
	if l < 1 {
		return nil, false
	}
	return b.entities[l-1], true
}
