package bin

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// error code:int32 message:string = Error;
type Message struct {
	Code    int32
	Message string
}

// EncodeTo implements bin.Encoder.
func (m Message) EncodeTo(b *Buffer) {
	b.PutID(0x9bdd8f1a)
	b.PutInt32(m.Code)
	b.PutString(m.Message)
}

func TestEncodeMessage(t *testing.T) {
	m := Message{
		Code:    204,
		Message: "Wake up, Neo",
	}
	b := new(Buffer)
	m.EncodeTo(b)
	expected := []byte{
		// Type ID.
		0x1a, 0x8f, 0xdd, 0x9b,

		// Code as int32.
		204, 0x00, 0x00, 0x00,

		// String length.
		byte(len(m.Message)),

		// "Wake up, Neo" in hex.
		0x57, 0x61, 0x6b,
		0x65, 0x20, 0x75, 0x70,
		0x2c, 0x20, 0x4e, 0x65,
		0x6f, 0x00, 0x00, 0x00,
	}
	if !bytes.Equal(expected, b.buf) {
		t.Log(hex.Dump(b.buf))
	}
}
