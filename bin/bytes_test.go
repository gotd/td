package bin

import (
	"bytes"
	"io"
	"testing"
)

func TestBytesDecodeEncode(t *testing.T) {
	for _, b := range [][]byte{
		bytes.Repeat([]byte{1, 2, 3, 4}, 100),
		bytes.Repeat([]byte{1, 2, 3}, 102),
		bytes.Repeat([]byte{1, 2}, 103),
		bytes.Repeat([]byte{10}, 104),
		bytes.Repeat([]byte{6}, 105),
		[]byte("foo"),
		[]byte("b"),
		[]byte("ba"),
		[]byte("what are you doing?"),
		[]byte("кек"),
		{
			0x57, 0x61, 0x6b, 0x65,
			0x20, 0x75, 0x70, 0x2c,
			0x20, 0x4e, 0x65, 0x6f,
		},
	} {
		buf := encodeBytes(nil, b)
		if len(buf)%4 != 0 {
			t.Error("bad align")
		}
		n, v, err := decodeBytes(buf)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(v, b) {
			t.Errorf("mismatch: %d != %d", len(v), len(b))
		}
		if n == 0 {
			t.Error("zero bytes read return")
		}
	}
	t.Run("ErrUnexpectedEOF", func(t *testing.T) {
		if _, _, err := decodeBytes(encodeBytes(nil, []byte{1, 2, 3, 4})[:2]); err != io.ErrUnexpectedEOF {
			t.Fatal("error expected")
		}
		if _, _, err := decodeBytes(encodeBytes(nil, bytes.Repeat([]byte{1, 2, 3, 4}, 105))[:10]); err != io.ErrUnexpectedEOF {
			t.Fatal("error expected")
		}
	})
}
