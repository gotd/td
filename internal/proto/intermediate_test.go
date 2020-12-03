package proto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func BenchmarkWriteIntermediate(b *testing.B) {
	out := new(bytes.Buffer)
	buf := new(bin.Buffer)
	buf.PutString("Hello world")
	buf.PutString("Wake up")
	buf.PutString("Neo")

	b.ReportAllocs()
	b.SetBytes(int64(buf.Len() + 4))

	for i := 0; i < b.N; i++ {
		if err := WriteIntermediate(out, buf); err != nil {
			b.Fatal(err)
		}
		out.Reset()
	}
}

func BenchmarkReadIntermediate(b *testing.B) {
	out := new(bytes.Buffer)
	buf := new(bin.Buffer)
	buf.PutString("Hello world")
	buf.PutString("Wake up")
	buf.PutString("Neo")
	if err := WriteIntermediate(out, buf); err != nil {
		b.Fatal(err)
	}
	raw := out.Bytes()
	reader := bytes.NewReader(nil)

	b.ReportAllocs()
	b.SetBytes(int64(buf.Len() + 4))

	for i := 0; i < b.N; i++ {
		reader.Reset(raw)
		if err := ReadIntermediate(reader, buf); err != nil {
			b.Fatal(err)
		}
		buf.Reset()
	}
}

func TestIntermediate(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		msg := bytes.Repeat([]byte{1, 2, 3}, 100)
		buf := new(bytes.Buffer)
		if err := WriteIntermediate(buf, &bin.Buffer{Buf: msg}); err != nil {
			t.Fatal(err)
		}
		var out bin.Buffer
		if err := ReadIntermediate(buf, &out); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, msg, out.Buf)
	})
	t.Run("BigMessage", func(t *testing.T) {
		t.Run("Read", func(t *testing.T) {
			var b bin.Buffer
			b.PutInt(1024*1024 + 10)

			var out bin.Buffer
			if err := ReadIntermediate(&b, &out); !errors.Is(err, ErrMessageTooBig) {
				t.Error(err)
			}
		})
		t.Run("Write", func(t *testing.T) {
			buf := make([]byte, 1024*1024*2)

			if err := WriteIntermediate(nil, &bin.Buffer{Buf: buf}); !errors.Is(err, ErrMessageTooBig) {
				t.Error(err)
			}
		})
	})
}
