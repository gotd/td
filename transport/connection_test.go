package transport

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/proto/codec"
)

func TestConnection(t *testing.T) {
	leftConn, rightConn := net.Pipe()
	intermediate := codec.Intermediate{}

	left := connection{
		conn:  leftConn,
		codec: intermediate,
	}
	right := connection{
		conn:  rightConn,
		codec: intermediate,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	buf := bytes.Repeat([]byte{1, 2, 3, 4}, 50)
	done := make(chan []byte)
	go func() {
		defer close(done)

		var b bin.Buffer
		if err := right.Recv(ctx, &b); err != nil {
			t.Error(err)
			return
		}

		done <- b.Buf
	}()

	if err := left.Send(ctx, &bin.Buffer{Buf: buf}); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, buf, <-done)
}
