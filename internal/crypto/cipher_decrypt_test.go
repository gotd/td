package crypto

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/testutil"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

func TestDecrypt(t *testing.T) {
	// Test vector from grammers.
	c := NewClientCipher(Zero{})
	var msg EncryptedMessage
	b := &bin.Buffer{Buf: []byte{
		122, 113, 131, 194, 193, 14, 79, 77, 249, 69, 250, 154, 154, 189, 53, 231, 195, 132,
		11, 97, 240, 69, 48, 79, 57, 103, 76, 25, 192, 226, 9, 120, 79, 80, 246, 34, 106, 7,
		53, 41, 214, 117, 201, 44, 191, 11, 250, 140, 153, 167, 155, 63, 57, 199, 42, 93, 154,
		2, 109, 67, 26, 183, 64, 124, 160, 78, 204, 85, 24, 125, 108, 69, 241, 120, 113, 82,
		78, 221, 144, 206, 160, 46, 215, 40, 225, 77, 124, 177, 138, 234, 42, 99, 97, 88, 240,
		148, 89, 169, 67, 119, 16, 216, 148, 199, 159, 54, 140, 78, 129, 100, 183, 100, 126,
		169, 134, 18, 174, 254, 148, 44, 93, 146, 18, 26, 203, 141, 176, 45, 204, 206, 182,
		109, 15, 135, 32, 172, 18, 160, 109, 176, 88, 43, 253, 149, 91, 227, 79, 54, 81, 24,
		227, 186, 184, 205, 8, 12, 230, 180, 91, 40, 234, 197, 109, 205, 42, 41, 55, 78,
	}}
	if err := msg.Decode(b); err != nil {
		t.Fatal(err)
	}
	plaintext, err := c.decryptMessage(testAuthKey, &msg)
	if err != nil {
		t.Fatal(err)
	}
	expectedPlaintext := []byte{
		252, 130, 106, 2, 36, 139, 40, 253, 96, 242, 196, 130, 36, 67, 173, 104, 1, 240, 193,
		194, 145, 139, 48, 94, 2, 0, 0, 0, 88, 0, 0, 0, 220, 248, 241, 115, 2, 0, 0, 0, 1, 168,
		193, 194, 145, 139, 48, 94, 1, 0, 0, 0, 28, 0, 0, 0, 8, 9, 194, 158, 196, 253, 51, 173,
		145, 139, 48, 94, 24, 168, 142, 166, 7, 238, 88, 22, 252, 130, 106, 2, 36, 139, 40,
		253, 1, 204, 193, 194, 145, 139, 48, 94, 2, 0, 0, 0, 20, 0, 0, 0, 197, 115, 119, 52,
		196, 253, 51, 173, 145, 139, 48, 94, 100, 8, 48, 0, 0, 0, 0, 0, 252, 230, 103, 4, 163,
		205, 142, 233, 208, 174, 111, 171, 103, 44, 96, 192, 74, 63, 31, 212, 73, 14, 81, 246,
	}
	if !bytes.Equal(expectedPlaintext, plaintext) {
		t.Error("mismatch")
	}
}

func TestCipher_Decrypt(t *testing.T) {
	var key AuthKey
	if _, err := io.ReadFull(testutil.Rand([]byte{10}), key.Value[:]); err != nil {
		t.Fatal(err)
	}

	c := NewClientCipher(Zero{})
	s := NewServerCipher(Zero{})
	tests := []struct {
		name      string
		data      []byte
		dataLen   int
		expectErr bool
	}{
		{"NegativeLength", []byte{1, 2, 3, 4}, -1, true},
		{"NoPadBy4", []byte{1, 2, 3}, 3, true},
		{"Good", bytes.Repeat([]byte{1, 2, 3, 4}, 4), 16, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)
			b := bin.Buffer{}
			data := EncryptedMessageData{
				MessageDataLen:         int32(test.dataLen),
				MessageDataWithPadding: test.data,
			}
			a.NoError(s.Encrypt(key, data, &b))

			_, err := c.DecryptFromBuffer(key, &b)
			if test.expectErr {
				a.Error(err)
				return
			}
			a.NoError(err)
		})
	}
}
