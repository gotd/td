package proto

import (
	"bytes"
	"testing"

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
