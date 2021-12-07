package entity

import (
	"io"
	"unicode/utf8"
)

// ComputeLength returns length of s encoded as UTF-16 string.
//
// While Telegram API docs state that they expect the number of UTF-8
// code points, in fact they are talking about UTF-16 code units.
func ComputeLength(s string) int {
	// From utf16 package.
	n := 0
	for _, v := range s {
		n += utf16RuneLen(v)
	}
	return n
}

// ComputeLengthBytes returns length of s encoded as UTF-16 string.
//
// While Telegram API docs state that they expect the number of UTF-8
// code points, in fact they are talking about UTF-16 code units.
func ComputeLengthBytes(s []byte) (n int) {
	// From utf16 package.
	var i int
	for i < len(s) {
		v, size := utf8.DecodeRune(s[i:])
		i += size
		n += utf16RuneLen(v)
	}
	return n
}

func utf16RuneLen(v rune) int {
	const (
		surrSelf = 0x10000
		maxRune  = '\U0010FFFF' // Maximum valid Unicode code point.
	)

	if surrSelf <= v && v <= maxRune {
		return 2
	}
	return 1
}

func (b *Builder) appendMessage(s string, formats ...Formatter) *Builder {
	if s == "" {
		return b
	}

	offset := b.utf16length
	length := ComputeLength(s)

	b.appendEntities(offset, length, utf8entity{
		offset: b.message.Len(),
		length: len(s),
	}, formats...)
	_, _ = b.WriteString(s)
	return b
}

func (b *Builder) appendEntities(offset, length int, u utf8entity, formats ...Formatter) *Builder {
	b.lastFormatIndex = len(b.entities)
	for i := range formats {
		b.entities = append(b.entities, formats[i](offset, length))
		b.lengths = append(b.lengths, u)
	}
	return b
}

var _ = []interface {
	io.Writer
	io.StringWriter
	io.ByteWriter
	WriteRune(rune) (int, error)
}{
	(*Builder)(nil),
}

// Write implements io.Writer.
func (b *Builder) Write(s []byte) (int, error) {
	n, err := b.message.Write(s)
	b.utf16length += ComputeLengthBytes(s)
	return n, err
}

// WriteString implements io.StringWriter.
func (b *Builder) WriteString(s string) (int, error) {
	n, err := b.message.WriteString(s)
	b.utf16length += ComputeLength(s)
	return n, err
}

// WriteByte implements io.ByteWriter.
func (b *Builder) WriteByte(s byte) error {
	err := b.message.WriteByte(s)
	b.utf16length++
	return err
}

// WriteRune implements rune writer.
func (b *Builder) WriteRune(s rune) (int, error) {
	n, err := b.message.WriteRune(s)
	b.utf16length += utf16RuneLen(s)
	return n, err
}
