package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto/codec"
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

func TestConnectionPartialFrameBytesRead(t *testing.T) {
	leftConn, rightConn := net.Pipe()
	t.Cleanup(func() {
		_ = leftConn.Close()
		_ = rightConn.Close()
	})

	payload := bytes.Repeat([]byte{1, 2, 3, 4}, 50)
	var length [4]byte
	binary.LittleEndian.PutUint32(length[:], uint32(len(payload)))
	full := uint64(len(length) + len(payload))

	right := &connection{
		conn:  rightConn,
		codec: codec.Intermediate{},
	}

	errc := make(chan error, 1)
	go func() {
		var b bin.Buffer
		errc <- right.Recv(context.Background(), &b)
	}()

	// Half of the length prefix. Recv stays blocked inside ReadFull, but the
	// bytes already pulled off the pipe must count.
	_, err := leftConn.Write(length[:2])
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return right.BytesRead() > 0
	}, time.Second, time.Millisecond)
	require.Less(t, right.BytesRead(), full)

	_, err = leftConn.Write(length[2:])
	require.NoError(t, err)
	_, err = leftConn.Write(payload)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return right.BytesRead() == full
	}, time.Second, time.Millisecond)

	select {
	case err := <-errc:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Recv did not return")
	}
	require.Equal(t, full, right.BytesRead())
}
