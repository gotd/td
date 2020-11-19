package td

import (
	"testing"

	"github.com/ernado/td/internal/bin"
)

func BenchmarkMessage_Encode(b *testing.B) {
	b.ReportAllocs()

	buf := new(bin.Buffer)
	msg := Message{
		Err: Error{
			Message:   "Foo",
			Code:      134,
			Temporary: true,
		},
	}
	for i := 0; i < b.N; i++ {
		msg.Encode(buf)
		buf.Reset()
	}
}

func BenchmarkMessage_Decode(b *testing.B) {
	b.ReportAllocs()

	encodeBuf := new(bin.Buffer)
	msg := &Message{
		Err: Error{
			Message:   "Foo",
			Code:      134,
			Temporary: true,
		},
	}
	msg.Encode(encodeBuf)
	raw := encodeBuf.Bytes()
	b.SetBytes(int64(len(raw)))

	buf := new(bin.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded Message
		buf.ResetTo(raw)
		if err := decoded.Decode(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func TestMessage(t *testing.T) {
	b := new(bin.Buffer)
	msg := Message{
		Err: Error{
			Message:   "Foo",
			Code:      134,
			Temporary: true,
		},
	}
	msg.Encode(b)

	result := Message{}
	if err := result.Decode(b); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkDecodeBool(b *testing.B) {
	b.ReportAllocs()

	encodeBuf := new(bin.Buffer)
	BoolTrue{}.Encode(encodeBuf)
	raw := encodeBuf.Bytes()
	b.SetBytes(int64(len(raw)))

	buf := new(bin.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.ResetTo(raw)
		v, err := DecodeBool(buf)
		if err != nil {
			b.Fatal(err)
		}
		switch v.(type) {
		case *BoolTrue: // ok
		default:
			b.Fatalf("Unexpected %T", v)
		}
	}
}

func BenchmarkDecodeResponse(b *testing.B) {
	b.ReportAllocs()

	encodeBuf := new(bin.Buffer)
	ResponseID{ID: 13}.Encode(encodeBuf)
	raw := encodeBuf.Bytes()
	b.SetBytes(int64(len(raw)))

	buf := new(bin.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.ResetTo(raw)
		v, err := DecodeResponse(buf)
		if err != nil {
			b.Fatal(err)
		}
		switch v.(type) {
		case *ResponseID: // ok
		default:
			b.Fatalf("Unexpected %T", v)
		}
	}
}

func BenchmarkDecodeAbstractMessage(b *testing.B) {
	b.Run("NoMessage", func(b *testing.B) {
		b.ReportAllocs()

		encodeBuf := new(bin.Buffer)
		NoMessage{}.Encode(encodeBuf)
		raw := encodeBuf.Bytes()
		b.SetBytes(int64(len(raw)))

		buf := new(bin.Buffer)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf.ResetTo(raw)
			v, err := DecodeAbstractMessage(buf)
			if err != nil {
				b.Fatal(err)
			}
			switch v.(type) {
			case *NoMessage: // ok
			default:
				b.Fatalf("Unexpected %T", v)
			}
		}
	})
	b.Run("BigMessage", func(b *testing.B) {
		b.ReportAllocs()

		encodeBuf := new(bin.Buffer)
		BigMessage{}.Encode(encodeBuf)
		raw := encodeBuf.Bytes()
		b.SetBytes(int64(len(raw)))

		buf := new(bin.Buffer)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf.ResetTo(raw)
			v, err := DecodeAbstractMessage(buf)
			if err != nil {
				b.Fatal(err)
			}
			switch v.(type) {
			case *BigMessage: // ok
			default:
				b.Fatalf("Unexpected %T", v)
			}
		}
	})
}
