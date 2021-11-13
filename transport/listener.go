package transport

import (
	"io"
	"net"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
)

// Listener is a simple net.Listener wrapper for listening
// MTProto transport connections.
type Listener struct {
	codec    func() Codec
	listener net.Listener
}

// Listen creates new Listener using given net.Listener.
// Transport codec will be detected automatically.
func Listen(listener net.Listener) Listener {
	return ListenCodec(nil, listener)
}

// ListenCodec creates new Listener using given net.Listener.
// Listener will always use given Codec constructor.
func ListenCodec(codec func() Codec, listener net.Listener) Listener {
	return Listener{
		codec:    codec,
		listener: &onceCloseListener{Listener: listener},
	}
}

type wrappedConn struct {
	reader io.Reader
	net.Conn
}

func (w wrappedConn) Read(b []byte) (int, error) {
	return w.reader.Read(b)
}

// Accept waits for and returns the next connection to the listener.
func (l Listener) Accept() (_ Conn, rErr error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	defer func() {
		if rErr != nil {
			multierr.AppendInto(&rErr, conn.Close())
		}
	}()

	// If codec provided explicitly, use it.
	if l.codec != nil {
		codec := l.codec()

		if err := codec.ReadHeader(conn); err != nil {
			return nil, errors.Wrap(err, "read header")
		}

		return &connection{
			conn:  conn,
			codec: codec,
		}, nil
	}

	// Otherwise try to detect codec.
	transportCodec, reader, err := detectCodec(conn)
	if err != nil {
		return nil, errors.Wrap(err, "detect codec")
	}

	return &connection{
		conn: wrappedConn{
			reader: reader,
			Conn:   conn,
		},
		codec: transportCodec,
	}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l Listener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address.
func (l Listener) Addr() net.Addr {
	return l.listener.Addr()
}
