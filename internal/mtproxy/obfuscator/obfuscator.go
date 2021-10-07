// Package obfuscator contains some MTProxy obfuscation utilities.
package obfuscator

import (
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/mtproxy"
	"github.com/nnqq/td/internal/mtproxy/faketls"
	"github.com/nnqq/td/internal/mtproxy/obfuscated2"
)

// Obfuscator represents MTProxy obfuscator.
type Obfuscator interface {
	io.ReadWriter
	Handshake(protocol [4]byte, s mtproxy.Secret) error
}

type tls struct {
	ftls  *faketls.FakeTLS
	obfs2 *obfuscated2.Obfuscated2
}

func newTLS(rand io.Reader, conn io.ReadWriter) tls {
	ftls := faketls.NewFakeTLS(rand, conn)
	obfs2 := obfuscated2.NewObfuscated2(rand, ftls)
	return tls{
		ftls:  ftls,
		obfs2: obfs2,
	}
}

func (t tls) Write(p []byte) (int, error) {
	return t.obfs2.Write(p)
}

func (t tls) Read(p []byte) (int, error) {
	return t.obfs2.Read(p)
}

func (t tls) Handshake(protocol [4]byte, s mtproxy.Secret) error {
	if err := t.ftls.Handshake(protocol, s); err != nil {
		return xerrors.Errorf("faketls handshake: %w", err)
	}

	if err := t.obfs2.Handshake(protocol, s); err != nil {
		return xerrors.Errorf("obfs2 handshake: %w", err)
	}

	return nil
}
