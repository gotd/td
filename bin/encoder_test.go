package bin

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// error code:int32 message:string = Error;
type Message struct {
	Code    int32
	Message string
	Nonce   Int128
	Key     Int256
}

// EncodeTo implements bin.Encoder.
func (m Message) Encode(b *Buffer) error {
	b.PutID(0x9bdd8f1a)
	b.PutInt32(m.Code)
	b.PutString(m.Message)
	b.PutInt128(m.Nonce)
	b.PutInt256(m.Key)
	return nil
}

func (m *Message) Decode(b *Buffer) error {
	if err := b.ConsumeID(0x9bdd8f1a); err != nil {
		return err
	}
	{
		v, err := b.Int32()
		if err != nil {
			return err
		}
		m.Code = v
	}
	{
		v, err := b.String()
		if err != nil {
			return err
		}
		m.Message = v
	}
	{
		v, err := b.Int128()
		if err != nil {
			return err
		}
		m.Nonce = v
	}
	{
		v, err := b.Int256()
		if err != nil {
			return err
		}
		m.Key = v
	}
	return nil
}

func TestEncodeMessage(t *testing.T) {
	m := Message{
		Code:    204,
		Message: "Wake up, Neo",
		Nonce:   [16]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF},
		Key: [32]byte{
			0xFF, 0xAA, 0xFF, 0xBB, 0xEE, 0x11, 0x12, 0x13, 0x14, 0x10, 0x10, 0x02, 0x04, 0x06, 0x08, 0x0A,
			0x00, 0x00, 0x00, 0x33, 0x55, 0xEE, 0x16, 0x11, 0x10, 0x14, 0x15, 0x02, 0x10, 0x10, 0x20, 0x20,
		},
	}
	b := new(Buffer)
	_ = m.Encode(b)
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

		// Nonce.
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,

		// Key.
		0xFF, 0xAA, 0xFF, 0xBB, 0xEE, 0x11, 0x12, 0x13,
		0x14, 0x10, 0x10, 0x02, 0x04, 0x06, 0x08, 0x0A,
		0x00, 0x00, 0x00, 0x33, 0x55, 0xEE, 0x16, 0x11,
		0x10, 0x14, 0x15, 0x02, 0x10, 0x10, 0x20, 0x20,
	}
	if !bytes.Equal(expected, b.Buf) {
		t.Log(hex.Dump(b.Buf))
	}
	var decoded Message
	if err := decoded.Decode(b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, m, decoded)

	t.Run("BigInt", func(t *testing.T) {
		// Byte order should not change (as defined in Telegram docs) and should be
		// "same as openssl".
		// This rule is correct for cryptographical big integers and parts of SHA1-s.
		expectedKey, ok := big.NewInt(0).
			SetString("FFAAFFBBEE111213141010020406080a0000003355EE16111014150210102020", 16)
		require.True(t, ok)
		require.Zero(t, expectedKey.Cmp(decoded.Key.BigInt()), "key big.Int unexpected")

		expectedNonce, ok := big.NewInt(0).SetString("000102030405060708090A0B0C0D0E0F", 16)
		require.True(t, ok)
		require.Zero(t, expectedNonce.Cmp(decoded.Nonce.BigInt()), "nonce big.Int unexpected")
	})
}
