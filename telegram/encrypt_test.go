package telegram

import (
	"bytes"
	"testing"

	"github.com/ernado/td/bin"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

func TestEncryption(t *testing.T) {
	c := &Client{
		rand: Zero{},
	}
	for i := 0; i < 256; i++ {
		c.authKey[i] = byte(i)
	}
	c.authKeyID = c.authKey.ID()

	// Testing vector from grammers.
	msg, err := c.encrypt([]byte("Hello, world! This data should remain secure!"))
	if err != nil {
		t.Fatal(err)
	}
	b := &bin.Buffer{}
	if err := msg.Encode(b); err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		50, 209, 88, 110, 164, 87, 223, 200, 168, 23, 41, 212, 109, 181, 64, 25, 162, 191, 215,
		247, 68, 249, 185, 108, 79, 113, 108, 253, 196, 71, 125, 178, 162, 193, 95, 109, 219,
		133, 35, 95, 185, 85, 47, 29, 132, 7, 198, 170, 234, 0, 204, 132, 76, 90, 27, 246, 172,
		68, 183, 155, 94, 220, 42, 35, 134, 139, 61, 96, 115, 165, 144, 153, 44, 15, 41, 117,
		36, 61, 86, 62, 161, 128, 210, 24, 238, 117, 124, 154,
	}
	if !bytes.Equal(b.Buf, expected) {
		t.Error("mismatch")
	}
}
