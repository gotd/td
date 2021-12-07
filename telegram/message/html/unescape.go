package html

import "unicode/utf8"

// unescapeEntity reads an entity like "&lt;" from b[src:] and writes the
// corresponding "<" to b[dst:], returning the incremented dst and src cursors.
// Precondition: b[src] == '&' && dst <= src.
//
// This is adaption of html.UnescapeString from Go sources.
func unescapeEntity(b []byte, dst, src int) (dst1, src1 int) {
	// i starts at 1 because we already know that s[0] == '&'.
	i, s := 1, b[src:]

	if len(s) <= 1 {
		b[dst] = b[src]
		return dst + 1, src + 1
	}

	if s[i] == '#' {
		if len(s) <= 3 { // We need to have at least "&#.".
			b[dst] = b[src]
			return dst + 1, src + 1
		}
		i++
		c := s[i]
		hex := false
		if c == 'x' || c == 'X' {
			hex = true
			i++
		}

		x := '\x00'
		for i < len(s) {
			c = s[i]
			i++
			if hex {
				switch {
				case '0' <= c && c <= '9':
					x = 16*x + rune(c) - '0'
					continue
				case 'a' <= c && c <= 'f':
					x = 16*x + rune(c) - 'a' + 10
					continue
				case 'A' <= c && c <= 'F':
					x = 16*x + rune(c) - 'A' + 10
					continue
				}
			} else if '0' <= c && c <= '9' {
				x = 10*x + rune(c) - '0'
				continue
			}
			if c != ';' {
				i--
			}
			break
		}

		if i <= 3 { // No characters matched.
			b[dst] = b[src]
			return dst + 1, src + 1
		}

		if x == 0 || x >= 0x10ffff {
			b[dst] = b[src]
			return dst + 1, src + 1
		}

		return dst + utf8.EncodeRune(b[dst:], x), src + i
	}

	// Consume the maximum number of characters possible, with the
	// consumed characters matching one of the named references.

	for i < len(s) {
		c := s[i]
		i++
		// Lower-cased characters are more common in entities, so we check for them first.
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
			continue
		}
		if c != ';' {
			i--
		}
		break
	}

	var x rune
	tagEnd := i
	if i > 0 && s[tagEnd-1] == ';' {
		tagEnd--
	}
	switch string(s[1:tagEnd]) {
	case "lt":
		x = '<'
	case "gt":
		x = '>'
	case "amp":
		x = '&'
	case "quot":
		x = '"'
	}
	if x != 0 {
		return dst + utf8.EncodeRune(b[dst:], x), src + i
	}

	dst1, src1 = dst+i, src+i
	copy(b[dst:dst1], b[src:src1])
	return dst1, src1
}

// telegramEscape implements Telegram BotAPI HTML unescape.
func telegramUnescape(b []byte) []byte {
	for i, c := range b {
		if c == '&' {
			dst, src := unescapeEntity(b, i, i)
			for src < len(b) {
				c := b[src]
				if c == '&' {
					dst, src = unescapeEntity(b, dst, src)
				} else {
					b[dst] = c
					dst, src = dst+1, src+1
				}
			}
			return b[0:dst]
		}
	}
	return b
}
