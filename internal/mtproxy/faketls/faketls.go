package faketls

import (
	"bytes"
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/mtproxy"
)

// FakeTLS implements FakeTLS obfuscation protocol.
type FakeTLS struct {
	rand io.Reader
	conn io.ReadWriter

	version [2]byte
	buf     bytes.Buffer
}

// NewFakeTLS creates new FakeTLS.
func NewFakeTLS(r io.Reader, conn io.ReadWriter) *FakeTLS {
	return &FakeTLS{
		rand:    r,
		conn:    conn,
		version: Version10Bytes,
		buf:     bytes.Buffer{},
	}
}

// Handshake performs FakeTLS handshake.
func (o *FakeTLS) Handshake(protocol [4]byte, s mtproxy.Secret) error {
	o.buf.Reset()

	var sessionID [32]byte
	if _, err := o.rand.Read(sessionID[:]); err != nil {
		return xerrors.Errorf("generate sessionID: %w", err)
	}

	clientDigest, err := writeClientHello(o.conn, sessionID, s.Secret)
	if err != nil {
		return xerrors.Errorf("send ClientHello: %w", err)
	}

	if err := readServerHello(o.conn, clientDigest, s.Secret); err != nil {
		return xerrors.Errorf("receive ServerHello: %w", err)
	}

	return nil
}

// Write implements io.Writer.
func (o *FakeTLS) Write(b []byte) (n int, err error) {
	n, err = writeRecord(o.conn, record{
		Type:    RecordTypeApplication,
		Version: o.version,
		Data:    b,
	})
	if err != nil {
		return 0, xerrors.Errorf("write TLS record: %w", err)
	}
	return
}

// Read implements io.Reader.
func (o *FakeTLS) Read(b []byte) (n int, err error) {
	if o.buf.Len() > 0 {
		return o.buf.Read(b)
	}

	rec, err := readRecord(o.conn)
	if err != nil {
		return 0, xerrors.Errorf("read TLS record: %w", err)
	}

	switch rec.Type {
	case RecordTypeChangeCipherSpec:
	case RecordTypeApplication:
	case RecordTypeHandshake:
		return 0, xerrors.New("unexpected record type handshake")
	default:
		return 0, xerrors.Errorf("unsupported record type %v", rec.Type)
	}
	o.buf.Write(rec.Data)

	return o.buf.Read(b)
}
