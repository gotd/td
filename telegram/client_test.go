package telegram

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ernado/td/internal/proto"
)

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func TestClient_Connect(t *testing.T) {
	listener := newLocalListener()
	defer func() { _ = listener.Close() }()

	ctx := context.Background()
	client, err := Dial(ctx, Options{
		Addr:    listener.Addr().String(),
		Network: listener.Addr().Network(),
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	read := false
	go func() {
		defer close(done)
		con, err := listener.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 1024)
		n, err := con.Read(buf)
		if err != nil {
			return
		}
		buf = buf[:n]
		read = bytes.Equal(buf, proto.IntermediateClientStart)
	}()

	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
		require.True(t, read, "server side should be read")
	case <-time.After(time.Second * 10):
		t.Fatal("timed out")
	}
}
