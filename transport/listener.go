package transport

import (
	"io"
	"net"

	"golang.org/x/xerrors"
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
	return Listener{
		listener: &onceCloseListener{Listener: listener},
	}
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
func (l Listener) Accept() (Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	// If codec provided explicitly, use it.
	if l.codec != nil {
		codec := l.codec()

		if err := codec.ReadHeader(conn); err != nil {
			return nil, xerrors.Errorf("read header: %w", err)
		}

		return &connection{
			conn:  conn,
			codec: codec,
		}, nil
	}

	// Otherwise try to detect codec.
	transportCodec, reader, err := detectCodec(conn)
	if err != nil {
		return nil, xerrors.Errorf("detect codec: %w", err)
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
