package entity

import "github.com/gotd/td/tg"

// Formatter is a message entity constructor.
type Formatter func(offset, limit int) tg.MessageEntityClass

// Plain formats message as plain text.
func (b *Builder) Plain(s string) *Builder {
	_, _ = b.WriteString(s)
	b.lastFormatIndex = len(b.entities)
	return b
}

// Format formats message using given formatters.
func (b *Builder) Format(s string, formats ...Formatter) *Builder {
	return b.appendMessage(s, formats...)
}

//go:generate go run github.com/gotd/td/telegram/message/internal/mkentity -output options.gen.go
