// Package keyparser extracts public keys from code.
package keyparser

import (
	"bufio"
	"io"
	"strings"
)

const (
	rsaBegin = "BEGIN"
	rsaEnd   = "END"
	rsaSep   = "-----"
)

func isBegin(s string) bool {
	if !strings.Contains(s, rsaSep) {
		return false
	}
	return strings.Contains(s, rsaBegin)
}

func isEnd(s string) bool {
	if !strings.Contains(s, rsaSep) {
		return false
	}
	return strings.Contains(s, rsaEnd)
}

// cleanHeader cleans out any suffix from header.
//
// If s is not header, s is returned unchanged.
func cleanHeader(s string) string {
	if !strings.HasPrefix(s, rsaSep) {
		return s
	}
	idx := strings.LastIndex(s, rsaSep)
	if idx <= 0 {
		return s
	}

	return s[:idx+len(rsaSep)]
}

// Extract public keys from C++ code in r to w.
func Extract(r io.Reader, w io.Writer) error {
	s := bufio.NewScanner(r)

	var (
		parts []string
		body  bool
		b     strings.Builder
	)

	for s.Scan() {
		text := strings.TrimSuffix(
			strings.TrimSpace(s.Text()), `\n\`,
		)
		text = strings.Trim(text, `"`)
		text = strings.TrimSuffix(text, `\n`)

		text = cleanHeader(text)

		if isBegin(text) {
			// Public key started.
			body = true
		}

		if body {
			parts = append(parts, text)
		}

		if isEnd(text) {
			// Public key completed.
			// Writing single public key to w.
			for _, part := range parts {
				b.WriteString(part)
				b.WriteRune('\n')
			}
			if _, err := io.WriteString(w, b.String()); err != nil {
				return err
			}

			// Reset state.
			b.Reset()
			body = false
			parts = parts[:0]
		}
	}

	if s.Err() != nil {
		return s.Err()
	}

	return nil
}
