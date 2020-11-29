package telegram

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"
)

func TestEncryption(t *testing.T) {
	c := &Client{
		rand: Zero{},
		log:  zap.NewNop(),
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

type testPayload struct {
	Size int
}

func (t testPayload) Encode(b *bin.Buffer) error {
	b.Buf = append(b.Buf, make([]byte, t.Size)...)
	return nil
}

func benchPayload(b *testing.B, c *Client, n int) {
	b.Helper()

	now := time.Date(1984, 10, 10, 0, 1, 2, 1249, time.UTC)

	buf := new(bin.Buffer)
	p := testPayload{Size: n}
	if err := c.newEncryptedMessage(crypto.NewMessageID(now, crypto.MessageFromClient), 0, p, buf); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(buf.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		id := crypto.NewMessageID(now, crypto.MessageFromClient)
		if err := c.newEncryptedMessage(id, 0, p, buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryption(b *testing.B) {
	c := &Client{
		rand: Zero{},
		log:  zap.NewNop(),
	}
	for i := 0; i < 256; i++ {
		c.authKey[i] = byte(i)
	}
	c.authKeyID = c.authKey.ID()

	for _, payload := range []int{
		128,
		1024,
		5000,
	} {
		b.Run(fmt.Sprintf("%d", payload), func(b *testing.B) {
			benchPayload(b, c, payload)
		})
	}
}
