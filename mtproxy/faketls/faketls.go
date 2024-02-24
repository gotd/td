package faketls

import (
	"bytes"
	"io"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/mtproxy"
)

// FakeTLS implements FakeTLS obfuscation protocol.
type FakeTLS struct {
	rand  io.Reader
	clock clock.Clock
	conn  io.ReadWriter

	version     [2]byte
	firstPacket bool

	readBuf    bytes.Buffer
	readBufMux sync.Mutex
}

// NewFakeTLS creates new FakeTLS.
func NewFakeTLS(r io.Reader, conn io.ReadWriter) *FakeTLS {
	return &FakeTLS{
		rand:    r,
		clock:   clock.System,
		conn:    conn,
		version: Version12Bytes,
		readBuf: bytes.Buffer{},
	}
}

// Handshake performs FakeTLS handshake.
func (o *FakeTLS) Handshake(protocol [4]byte, dc int, s mtproxy.Secret) error {
	o.readBufMux.Lock()
	o.readBuf.Reset()
	o.readBufMux.Unlock()

	var sessionID [32]byte
	if _, err := o.rand.Read(sessionID[:]); err != nil {
		return errors.Wrap(err, "generate sessionID")
	}

	clientDigest, err := writeClientHello(o.conn, o.clock, sessionID, s.CloakHost, s.Secret)
	if err != nil {
		return errors.Wrap(err, "send ClientHello")
	}

	if err := readServerHello(o.conn, clientDigest, s.Secret); err != nil {
		return errors.Wrap(err, "receive ServerHello")
	}
	return nil
}

// Write implements io.Writer.
func (o *FakeTLS) Write(b []byte) (n int, err error) {
	// Some proxies require sending first packet with RecordTypeChangeCipherSpec.
	//
	// See https://github.com/tdlib/td/blob/master/td/mtproto/TcpTransport.cpp#L266.
	if !o.firstPacket {
		_, err = writeRecord(o.conn, record{
			Type:    RecordTypeChangeCipherSpec,
			Version: o.version,
			Data:    []byte("\x01"),
		})
		if err != nil {
			return 0, errors.Wrap(err, "write first TLS packet")
		}
		o.firstPacket = true
	}
	n, err = writeRecord(o.conn, record{
		Type:    RecordTypeApplication,
		Version: o.version,
		Data:    b,
	})
	if err != nil {
		return 0, errors.Wrap(err, "write TLS record")
	}
	return
}

// Read implements io.Reader.
func (o *FakeTLS) Read(b []byte) (n int, err error) {
	o.readBufMux.Lock()
	defer o.readBufMux.Unlock()

	if o.readBuf.Len() > 0 {
		return o.readBuf.Read(b)
	}

	rec, err := readRecord(o.conn)
	if err != nil {
		return 0, errors.Wrap(err, "read TLS record")
	}

	switch rec.Type {
	case RecordTypeChangeCipherSpec:
	case RecordTypeApplication:
	case RecordTypeHandshake:
		return 0, errors.New("unexpected record type handshake")
	default:
		return 0, errors.Errorf("unsupported record type %v", rec.Type)
	}
	o.readBuf.Write(rec.Data)

	return o.readBuf.Read(b)
}
