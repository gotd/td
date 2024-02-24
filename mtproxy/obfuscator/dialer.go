package obfuscator

import (
	"io"
	"net"

	"github.com/gotd/td/mtproxy/obfuscated2"
)

// Conn is net.Conn wrapper to use Obfuscator.
type Conn struct {
	net.Conn
	Obfuscator
}

// Obfuscated2 creates new obfuscated2 connection.
func Obfuscated2(rand io.Reader, conn net.Conn) *Conn {
	return &Conn{
		Conn:       conn,
		Obfuscator: obfuscated2.NewObfuscated2(rand, conn),
	}
}

// FakeTLS creates new FakeTLS connection.
func FakeTLS(rand io.Reader, conn net.Conn) *Conn {
	return &Conn{
		Conn:       conn,
		Obfuscator: newTLS(rand, conn),
	}
}

// Write implements io.Writer.
func (o *Conn) Write(b []byte) (n int, err error) {
	return o.Obfuscator.Write(b)
}

// Read implements io.Reader.
func (o *Conn) Read(b []byte) (n int, err error) {
	return o.Obfuscator.Read(b)
}
