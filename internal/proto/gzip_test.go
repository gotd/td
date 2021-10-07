package proto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/testutil"
)

func TestGZIP_Encode(t *testing.T) {
	data := bytes.Repeat([]byte{1, 2, 3}, 100)
	g := &GZIP{
		Data: data,
	}

	var b bin.Buffer
	require.NoError(t, b.Encode(g))

	var decoded GZIP
	require.NoError(t, b.Decode(&decoded))
	require.Equal(t, data, decoded.Data)
}

func TestGZIP_Decode(t *testing.T) {
	g := &GZIP{
		Data: make([]byte, 1024*1024*15),
	}
	var b bin.Buffer
	require.NoError(t, b.Encode(g))

	var (
		decoded GZIP
		target  *DecompressionBombErr
	)
	require.ErrorAs(t, b.Decode(&decoded), &target)
	require.Less(t, len(decoded.Data), len(g.Data))
}

func benchmarkGZIPEncode(payloadSize int) func(b *testing.B) {
	return func(b *testing.B) {
		g := &GZIP{Data: make([]byte, payloadSize)}
		_, err := io.ReadFull(rand.Reader, g.Data)
		require.NoError(b, err)

		var buf bin.Buffer
		b.ReportAllocs()
		b.SetBytes(int64(payloadSize))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf.Reset()
			_ = g.Encode(&buf)
		}
	}
}

func BenchmarkGZIP_Encode(b *testing.B) {
	testutil.RunPayloads(b, benchmarkGZIPEncode)
}

func benchmarkGZIPDecode(payloadSize int) func(b *testing.B) {
	return func(b *testing.B) {
		g := &GZIP{Data: make([]byte, payloadSize)}
		_, err := io.ReadFull(rand.Reader, g.Data)
		require.NoError(b, err)

		var buf bin.Buffer
		require.NoError(b, g.Encode(&buf))
		b.ReportAllocs()
		b.SetBytes(int64(payloadSize))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = g.Decode(&bin.Buffer{Buf: buf.Buf})
		}
	}
}

func BenchmarkGZIP_Decode(b *testing.B) {
	testutil.RunPayloads(b, benchmarkGZIPDecode)
}
