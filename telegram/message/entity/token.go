package entity

// Token represents raw point in a message string.
type Token struct {
	utf8offset  int
	utf16offset int
}

// UTF8Offset return UTF-8 offset.
func (t Token) UTF8Offset() int {
	return t.utf8offset
}

// UTF16Offset returns UTF-16 offset.
func (t Token) UTF16Offset() int {
	return t.utf16offset
}

// UTF8Length return UTF-8 length between token start and current state.
func (t Token) UTF8Length(builder *Builder) int {
	return builder.UTF8Len() - t.utf8offset
}

// UTF16Length returns UTF-16 length between token start and current state.
func (t Token) UTF16Length(builder *Builder) int {
	return builder.UTF16Len() - t.utf16offset
}

// Text message string between token start and current state.
func (t Token) Text(builder *Builder) string {
	return builder.TextRange(t.utf8offset, builder.UTF8Len())
}

// Apply formats range between token start and current state using given Formatter slice.
func (t Token) Apply(builder *Builder, f ...Formatter) {
	builder.appendEntities(t.utf16offset, t.UTF16Length(builder), utf8entity{
		offset: t.utf8offset,
		length: t.UTF8Length(builder),
	}, f...)
}

// Token creates new Token.
func (b *Builder) Token() Token {
	return Token{
		utf8offset:  b.UTF8Len(),
		utf16offset: b.UTF16Len(),
	}
}
