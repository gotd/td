package proto

import (
	"context"
	"fmt"
	"net"

	"github.com/gotd/td/bin"
)

// Dialer dials using a context.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Transport is MTProto transport protocol abstraction.
type Transport interface {
	// Dial sends protocol version.
	Dial(ctx context.Context, network, addr string) error

	// Send sends message from given buffer.
	Send(ctx context.Context, b *bin.Buffer) error

	// Recv fills buffer with received message.
	Recv(ctx context.Context, b *bin.Buffer) error

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error
}

// ProtocolErr represents protocol level error.
type ProtocolErr struct {
	Code int32
}

func (p ProtocolErr) Error() string {
	switch p.Code {
	case CodeAuthKeyNotFound:
		return "auth key not found"
	case CodeTransportFlood:
		return "transport flood"
	default:
		return fmt.Sprintf("protocol error %d", p.Code)
	}
}

func checkProtocolError(b *bin.Buffer) error {
	if b.Len() != bin.Word {
		return nil
	}
	code, err := b.Int32()
	if err != nil {
		return err
	}
	return &ProtocolErr{Code: -code}
}
