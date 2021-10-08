package entity

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
)

type utf8entity struct {
	offset int
	length  int
}

// Builder builds message string and text entities.
type Builder struct {
	entities []tg.MessageEntityClass
	// lengths stores offset/length data too, but in UTF-8 codepoints
	lengths  []utf8entity
	// We store index of first entity added at last Format call.
	// It needed to trim space in all entities of last text block.
	lastFormatIndex int
	message         strings.Builder
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

func (b *Builder) reset() {
	b.message.Reset()
	b.entities = nil
}

// Complete returns build result and resets builder.
func (b *Builder) Complete() (string, []tg.MessageEntityClass) {
	msg := b.message.String()
	entities := b.entities
	b.reset()

	// If there are no entities or last text block does not have entities,
	// so we just return built message.
	if len(b.lengths) == 0 || b.lastFormatIndex >= len(entities) {
		return msg, entities
	}

	// Since Telegram client does not handle space after formatted message
	// we should compute length of the last block to trim it.
	// Get first entity of last text block.
	entity := b.lengths[len(b.lengths)-1]
	offset := entity.offset
	length := entity.length
	// Get last text block.
	lastBlock := msg[offset:]
	// Trim this block.
	trimmed := strings.TrimRightFunc(lastBlock, unicode.IsSpace)

	// If there are a difference, we should change length of the all entities.
	if length >= len(lastBlock) && len(trimmed) != len(lastBlock) {
		length := ComputeLength(trimmed)
		for idx := range entities[b.lastFormatIndex:] {
			setLength(idx, length, entities[b.lastFormatIndex:])
		}
		return msg[:offset+len(trimmed)], entities
	}

	return msg, entities
}

// setLength sets Length field of entity.
func setLength(index, value int, slice []tg.MessageEntityClass) {
	reflect.ValueOf(&slice[index]).
		Elem().Elem().Elem().
		FieldByName("Length").
		SetInt(int64(value))
}

// ComputeLength returns length of s encoded as UTF-16 string.
//
// While Telegram API docs state that they expect the number of UTF-8
// code points, in fact they are talking about UTF-16 code units.
func ComputeLength(s string) int {
	const (
		surrSelf = 0x10000
		maxRune  = '\U0010FFFF' // Maximum valid Unicode code point.
	)

	// From utf16 package.
	n := 0
	for _, v := range s {
		if surrSelf <= v && v <= maxRune {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func (b *Builder) appendMessage(s string, formats ...Formatter) *Builder {
	if s == "" {
		return b
	}

	offset := ComputeLength(b.message.String())
	length := ComputeLength(s)

	b.appendEntities(offset, length, utf8entity{
		offset: b.message.Len(),
		length: len(s),
	}, formats...)
	b.message.WriteString(s)
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
