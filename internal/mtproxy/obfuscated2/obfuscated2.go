package obfuscated2

import (
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mtproxy"
)

// Obfuscated2 implements obfuscated2 obfuscation protocol.
type Obfuscated2 struct {
	rand io.Reader
	conn io.ReadWriter

	keys
}

// NewObfuscated2 creates new Obfuscated2.
func NewObfuscated2(r io.Reader, conn io.ReadWriter) *Obfuscated2 {
	return &Obfuscated2{
		rand: r,
		conn: conn,
	}
}

// Handshake sends obfuscated2 header.
func (o *Obfuscated2) Handshake(protocol [4]byte, s mtproxy.Secret) error {
	keys, err := generateKeys(o.rand, protocol, s.Secret, s.DC)
	if err != nil {
		return xerrors.Errorf("generate keys: %w", err)
	}
	o.keys = keys

	if _, err := o.conn.Write(o.header); err != nil {
		return xerrors.Errorf("write obfuscated header: %w", err)
	}

	return nil
}

// Write implements io.Writer.
func (o *Obfuscated2) Write(b []byte) (n int, err error) {
	o.encrypt.XORKeyStream(b, b)
	return o.conn.Write(b)
}

// Read implements io.Reader.
func (o *Obfuscated2) Read(b []byte) (n int, err error) {
	n, err = o.conn.Read(b)
	if err != nil {
		return
	}
	o.decrypt.XORKeyStream(b, b)
	return
}
