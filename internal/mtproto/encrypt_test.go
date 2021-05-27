package mtproto

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/neo"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

type testPayload struct {
	Size int
}

func (t testPayload) Encode(b *bin.Buffer) error {
	b.Buf = append(b.Buf, make([]byte, t.Size)...)
	return nil
}

func benchPayload(b *testing.B, c *Conn, n int) {
	b.Helper()

	buf := new(bin.Buffer)
	p := testPayload{Size: n}
	if err := c.newEncryptedMessage(12345, 0, p, buf); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(buf.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := c.newEncryptedMessage(12345, 0, p, buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryption(b *testing.B) {
	c := &Conn{
		rand:   Zero{},
		log:    zap.NewNop(),
		cipher: crypto.NewClientCipher(Zero{}),
		clock:  neo.NewTime(time.Now()),
	}
	for i := 0; i < 256; i++ {
		c.authKey.Value[i] = byte(i)
	}

	for _, payload := range []int{
		128,
		1024,
		16 * 1024,
		512 * 1024,
	} {
		b.Run(fmt.Sprintf("%d", payload), func(b *testing.B) {
			benchPayload(b, c, payload)
		})
	}
}
