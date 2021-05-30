package mtproto

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/neo"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/testutil"
)

func benchWrite(payloadSize int) func(b *testing.B) {
	return func(b *testing.B) {
		a := require.New(b)
		logger := zap.NewNop()
		random := rand.Reader
		c := neo.NewTime(time.Now())

		var key crypto.Key
		_, err := io.ReadFull(random, key[:])
		a.NoError(err)
		authKey := key.WithID()

		payload := make([]byte, payloadSize)
		_, err = io.ReadFull(random, payload)
		a.NoError(err)
		data := &testPayload{Data: payload}

		conn := Conn{
			conn:    &constantConn{},
			clock:   c,
			rand:    random,
			cipher:  crypto.NewClientCipher(random),
			log:     logger,
			authKey: authKey,
		}
		b.ResetTimer()
		b.ReportAllocs()
		b.SetBytes(int64(payloadSize))

		for i := 0; i < b.N; i++ {
			_ = conn.write(context.Background(), 1, 1, data)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	for _, size := range testutil.Payloads() {
		b.Run(fmt.Sprintf("%db", size), benchWrite(size))
	}
}
