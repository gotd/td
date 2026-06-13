package dcs

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/mtproxy/obfuscated2"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// TestPlainObfuscatedDirect verifies that PlainOptions.Obfuscated wraps a
// regular (not TCP-obfuscated-only) DC connection in Obfuscated2 using the
// configured codec's tag, matching how Telegram Desktop connects directly.
func TestPlainObfuscatedDirect(t *testing.T) {
	a := require.New(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	a.NoError(err)
	defer func() { _ = ln.Close() }()

	type result struct {
		meta obfuscated2.Metadata
		err  error
	}
	resultCh := make(chan result, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			resultCh <- result{err: err}
			return
		}
		defer func() { _ = conn.Close() }()
		_, meta, err := obfuscated2.Accept(conn, nil)
		resultCh <- result{meta: meta, err: err}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	r := Plain(PlainOptions{
		Protocol:   transport.Abridged,
		Obfuscated: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := r.Primary(ctx, 2, List{
		Options: []tg.DCOption{
			{ID: 2, IPAddress: addr.IP.String(), Port: addr.Port},
		},
	})
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	got := <-resultCh
	a.NoError(got.err)
	// The obfuscation header must carry the abridged codec tag, not a plaintext
	// codec prefix on the wire.
	a.Equal(codec.Abridged{}.ObfuscatedTag(), got.meta.Protocol)
	a.Equal(uint16(2), got.meta.DC)
}
