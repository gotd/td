package codec

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/testutil"
)

func benchWrite(codec Codec) func(payloadSize int) func(b *testing.B) {
	return func(payloadSize int) func(b *testing.B) {
		return func(b *testing.B) {
			buf := bin.Buffer{Buf: make([]byte, payloadSize)}
			if _, err := io.ReadFull(rand.Reader, buf.Buf); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.SetBytes(int64(buf.Len() + 4))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if err := codec.Write(io.Discard, &buf); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	b.Run("Abridged", func(b *testing.B) {
		testutil.RunPayloads(b, benchWrite(Abridged{}))
	})
	b.Run("Intermediate", func(b *testing.B) {
		testutil.RunPayloads(b, benchWrite(Intermediate{}))
	})
	b.Run("PaddedIntermediate", func(b *testing.B) {
		testutil.RunPayloads(b, benchWrite(PaddedIntermediate{}))
	})
}

func benchRead(codec Codec) func(payloadSize int) func(b *testing.B) {
	return func(payloadSize int) func(b *testing.B) {
		return func(b *testing.B) {
			buf := bin.Buffer{Buf: make([]byte, payloadSize)}
			if _, err := io.ReadFull(rand.Reader, buf.Buf); err != nil {
				b.Fatal(err)
			}

			out := new(bytes.Buffer)
			if err := codec.Write(out, &buf); err != nil {
				b.Fatal(err)
			}
			raw := out.Bytes()
			reader := bytes.NewReader(nil)

			b.ReportAllocs()
			b.SetBytes(int64(buf.Len() + 4))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				reader.Reset(raw)
				if err := codec.Read(reader, &buf); err != nil {
					b.Fatal(err)
				}
				buf.Reset()
			}
		}
	}
}

func BenchmarkRead(b *testing.B) {
	b.Run("Abridged", func(b *testing.B) {
		testutil.RunPayloads(b, benchRead(Abridged{}))
	})
	b.Run("Intermediate", func(b *testing.B) {
		testutil.RunPayloads(b, benchRead(Intermediate{}))
	})
	b.Run("PaddedIntermediate", func(b *testing.B) {
		testutil.RunPayloads(b, benchRead(PaddedIntermediate{}))
	})
}
