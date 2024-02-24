package transport

import (
	"bytes"
	"io"
	"net"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/mtproxy/obfuscated2"
	"github.com/gotd/td/proto/codec"
)

type obfListener struct {
	listener net.Listener
}

type obfConn struct {
	reader io.Reader
	writer io.Writer
	net.Conn
}

func (c *obfConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *obfConn) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

// ObfuscatedListener creates new obfuscated2 listener using given net.Listener.
//
// Useful for creating Telegram servers:
//
//	transport.Listen(transport.ObfuscatedListener(ln))
func ObfuscatedListener(listener net.Listener) net.Listener {
	return obfListener{listener: listener}
}

// Accept waits for and returns the next connection to the listener.
func (l obfListener) Accept() (_ net.Conn, err error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			multierr.AppendInto(&err, conn.Close())
		}
	}()

	rw, md, err := obfuscated2.Accept(conn, nil)
	if err != nil {
		return nil, errors.Wrap(err, "accept")
	}

	var tag *bytes.Reader
	if md.Protocol[0] == codec.AbridgedClientStart[0] {
		// Abridged sends only byte for tag.
		tag = bytes.NewReader(md.Protocol[:1])
	} else {
		tag = bytes.NewReader(md.Protocol[:])
	}

	accepted := &obfConn{
		reader: io.MultiReader(tag, rw),
		writer: rw,
		Conn:   conn,
	}

	return accepted, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l obfListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address.
func (l obfListener) Addr() net.Addr {
	return l.listener.Addr()
}
