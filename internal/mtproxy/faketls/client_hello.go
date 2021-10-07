package faketls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
)

// writeClientHello writes faketls ClientHello.
// See https://github.com/9seconds/mtg/blob/e075169dd4e9fc4c2b1453668f85f5099c4fb895/tlstypes/client_hello.go#L38-L86.
// See https://tools.ietf.org/html/rfc5246#section-7.4.1.1.
func writeClientHello(w io.Writer, sessionID [32]byte, secret []byte) (r [32]byte, err error) {
	// TODO(tdakkota): For now, this code does not use writeRecord, because it
	// 	hard to use due to buffering to compute HMAC.
	b := new(bin.Buffer)

	const helloPadding = 512
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], helloPadding)
	// See https://github.com/9seconds/mtg/blob/bfeded67ee6d1bcc4f04c2e33b765cc73ff5d1a5/faketls/consts.go#L18-L28.
	b.Put([]byte{
		// Record header.
		byte(RecordTypeHandshake),
		Version10Bytes[0], Version10Bytes[1],
		length[0], length[1], // equal to [0x2, 0x0] = BigEndian(512)
	})
	// Count padding from current point.
	padFrom := len(b.Buf)
	b.Put([]byte{
		// Record payload.
		// Put handshake_type.
		byte(HandshakeTypeClient),
		// Put payload_length.
		0x0, 0x1, 0xfc,
	})

	// Put protocol_version.
	b.Put(Version12Bytes[:])
	// Put random.
	// See https://github.com/9seconds/mtg/blob/bfeded67ee6d1bcc4f04c2e33b765cc73ff5d1a5/faketls/client_protocol.go#L77
	cur := len(b.Buf)
	b.Expand(32)
	// Put session_id_length.
	b.Put([]byte{32})
	// Put session_id.
	b.Put(sessionID[:])
	// Put cipher_suites.
	b.PutUint16(0)
	// Put compression_methods.
	b.Put([]byte{0})

	// Pad record payload to helloPadding.
	pad := helloPadding + padFrom
	if len(b.Buf) < pad {
		b.Expand(pad - len(b.Buf))
	}

	// https://github.com/tdlib/td/blob/27d3fdd09d90f6b77ecbcce50b1e86dc4b3dd366/td/mtproto/TlsInit.cpp#L380-L384
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(b.Buf); err != nil {
		return [32]byte{}, xerrors.Errorf("hmac write: %w", err)
	}

	s := mac.Sum(nil)
	copy(b.Buf[cur:cur+32], s)
	// Overwrite last 4 bytes using final := original ^ timestamp.
	old := binary.LittleEndian.Uint32(b.Buf[cur+28 : cur+32])
	old ^= uint32(time.Now().Unix())
	binary.LittleEndian.PutUint32(b.Buf[cur+28:cur+32], old)

	// Copy ClientRandom for later use.
	copy(r[:], b.Buf[cur:cur+32])
	_, err = w.Write(b.Buf)
	return r, err
}
