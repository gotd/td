package mtproto

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/neo"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/testutil"
)

func benchEncryption(b *testing.B, c *Conn, n int) {
	b.Helper()

	buf := &bin.Buffer{Buf: make([]byte, 0, n)}
	p := testPayload{Data: make([]byte, n-4)}
	b.ReportAllocs()
	b.SetBytes(int64(n))
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
		rand:              Zero{},
		log:               zap.NewNop(),
		cipher:            crypto.NewClientCipher(Zero{}),
		clock:             neo.NewTime(time.Now()),
		compressThreshold: -1,
	}
	for i := 0; i < 256; i++ {
		c.authKey.Value[i] = byte(i)
	}

	for _, payload := range testutil.Payloads() {
		b.Run(fmt.Sprintf("%db", payload), func(b *testing.B) {
			benchEncryption(b, c, payload)
		})
	}
}
