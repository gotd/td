package rich

import (
	"strings"

	"github.com/gotd/td/tg"
)

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

// collapseWS collapses runs of inline whitespace into single spaces, preserving
// a single leading and trailing space, following HTML inline whitespace rules.
// Whitespace-only input collapses to a single space; empty input stays empty.
func collapseWS(s string) string {
	var b strings.Builder
	pending := false
	started := false
	for _, r := range s {
		if isSpace(r) {
			pending = true
			continue
		}
		if pending {
			b.WriteByte(' ')
			pending = false
		}
		b.WriteRune(r)
		started = true
	}
	if pending && (started || b.Len() == 0 && s != "") {
		b.WriteByte(' ')
	}
	return b.String()
}

// isBlankPlain reports whether t is a plain text node containing only
// whitespace.
func isBlankPlain(t tg.RichTextClass) bool {
	p, ok := t.(*tg.TextPlain)
	return ok && strings.TrimSpace(p.Text) == ""
}

// trimZero drops leading and trailing whitespace-only plain nodes and trims the
// outer whitespace of the remaining edge plain nodes.
func trimZero(texts []tg.RichTextClass) []tg.RichTextClass {
	for len(texts) > 0 && isBlankPlain(texts[0]) {
		texts = texts[1:]
	}
	for len(texts) > 0 && isBlankPlain(texts[len(texts)-1]) {
		texts = texts[:len(texts)-1]
	}
	if len(texts) == 0 {
		return nil
	}
	if p, ok := texts[0].(*tg.TextPlain); ok {
		texts[0] = Plain(strings.TrimLeft(p.Text, " \t\r\n\f"))
	}
	if p, ok := texts[len(texts)-1].(*tg.TextPlain); ok {
		texts[len(texts)-1] = Plain(strings.TrimRight(p.Text, " \t\r\n\f"))
	}
	return texts
}

// trimInline trims whitespace around an inline run and reports whether any
// meaningful content remains.
func trimInline(texts []tg.RichTextClass) (tg.RichTextClass, bool) {
	texts = trimZero(texts)
	if len(texts) == 0 {
		return nil, false
	}
	return join(texts), true
}
