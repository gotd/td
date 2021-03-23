// Package keyparser extracts public keys from code.
package keyparser

import (
	"bufio"
	"io"
	"strings"
)

const (
	rsaBegin    = `-----BEGIN RSA PUBLIC KEY-----`
	rsaEnd      = `-----END RSA PUBLIC KEY-----`
	rsaBeginKey = "BEGIN"
	rsaEndKey   = "END"
	rsaSep      = "-----"
)

func isBegin(s string) bool {
	if !strings.Contains(s, rsaSep) {
		return false
	}
	return strings.Contains(s, rsaBeginKey)
}

func isEnd(s string) bool {
	if !strings.Contains(s, rsaSep) {
		return false
	}
	return strings.Contains(s, rsaEndKey)
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

		// Telegram uses both variants:
		// * BEGIN RSA PUBLIC KEY
		// * BEGIN PUBLIC KEY
		// Just normalize to first one for convenience.
		switch {
		case isBegin(text):
			text = rsaBegin
		case isEnd(text):
			text = rsaEnd
		}

		if text == rsaBegin {
			// Public key started.
			body = true
		}

		if body {
			parts = append(parts, text)
		}

		if text == rsaEnd {
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
