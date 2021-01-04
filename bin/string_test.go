package bin

import (
	"io"
	"strings"
	"testing"
)

func TestStringDecodeEncode(t *testing.T) {
	for _, s := range []string{
		strings.Repeat("abcd", 100),
		strings.Repeat("abc", 102),
		strings.Repeat("de", 103),
		strings.Repeat("z", 104),
		strings.Repeat("b", 105),
		"foo",
		"b",
		"ba",
		"what are you doing?",
		"кек",
		strings.Repeat("a", 253),
	} {
		buf := encodeString(nil, s)
		if len(buf)%4 != 0 {
			t.Error("bad align")
		}
		n, v, err := decodeString(buf)
		if err != nil {
			t.Error(err)
		}
		if v != s {
			t.Errorf("mismatch: %d != %d", len(v), len(s))
		}
		if n == 0 {
			t.Error("zero bytes read return")
		}
	}
	t.Run("ErrUnexpectedEOF", func(t *testing.T) {
		if _, _, err := decodeString(encodeString(nil, "foo bar")[:2]); err != io.ErrUnexpectedEOF {
			t.Fatal("error expected")
		}
		if _, _, err := decodeString(encodeString(nil, strings.Repeat("b", 105))[:10]); err != io.ErrUnexpectedEOF {
			t.Fatal("error expected")
		}
	})
}
