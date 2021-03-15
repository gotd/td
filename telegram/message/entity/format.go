package entity

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/gotd/td/tg"
)

// Builder builds message string and text entities.
type Builder struct {
	entities []tg.MessageEntityClass
	message  strings.Builder
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

	if len(entities) == 0 {
		return msg, nil
	}

	if len(entities) >= 1 {
		last := entities[len(entities)-1]
		offset := last.GetOffset()

		entityText := msg[offset:]
		trimmed := strings.TrimRightFunc(entityText, unicode.IsSpace)
		if len(trimmed) != len(entityText) {
			reflect.ValueOf(&entities[len(entities)-1]).
				Elem().Elem().Elem().
				FieldByName("Length").
				SetInt(int64(computeLength(trimmed)))
		}
	}

	return msg, entities
}

func computeLength(s string) int {
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

type formatter func(offset, limit int) tg.MessageEntityClass

func (b *Builder) appendMessage(s string, format formatter) *Builder {
	offset := b.message.Len()
	length := computeLength(s)

	b.entities = append(b.entities, format(offset, length))
	b.message.WriteString(s)
	return b
}
