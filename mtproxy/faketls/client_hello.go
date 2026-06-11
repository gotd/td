package faketls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"
	utls "github.com/refraction-networking/utls"

	"github.com/gotd/td/clock"
)

// clientRandomOffset is the offset of the 32-byte ClientRandom field inside a
// TLS ClientHello record:
//
//	5 bytes record header (type + version + length)
//	1 byte  handshake type
//	3 bytes handshake length
//	2 bytes client version
const (
	clientRandomOffset = 11
	clientRandomLength = 32
)

// generateClientHello builds a browser-like TLS ClientHello record using uTLS.
//
// Instead of hand-crafting the bytes (as older clients did), we mimic the
// fingerprint of a recent Chrome release so that the handshake is
// indistinguishable from a real browser connecting to the cloak domain. uTLS
// tracks browser fingerprints upstream, keeping us in line with the approach
// taken by the official Telegram clients.
//
// See https://github.com/refraction-networking/utls and
// https://github.com/tdlib/td/blob/master/td/mtproto/TlsInit.cpp.
func generateClientHello(rand io.Reader, domain string) ([]byte, error) {
	config := &utls.Config{
		ServerName: domain,
		// uTLS uses Rand for the ClientHello random, session ID and key shares.
		Rand: rand,
	}

	// The connection is never used: BuildHandshakeState only assembles the
	// ClientHello in memory, it does not perform any I/O.
	conn := utls.UClient(nil, config, utls.HelloChrome_Auto)
	if err := conn.BuildHandshakeState(); err != nil {
		return nil, errors.Wrap(err, "build handshake state")
	}

	hello := conn.HandshakeState.Hello.Raw
	// hello is the handshake message; wrap it into a TLS handshake record.
	record := make([]byte, 0, 5+len(hello))
	record = append(record, byte(RecordTypeHandshake), Version10Bytes[0], Version10Bytes[1])
	record = binary.BigEndian.AppendUint16(record, uint16(len(hello)))
	record = append(record, hello...)
	return record, nil
}

// writeClientHello writes faketls ClientHello.
//
// The ClientHello carries the MTProxy digest in its ClientRandom field: the
// HMAC-SHA256 of the whole record (computed with the ClientRandom zeroed) using
// the proxy secret as key, with the lower 4 bytes XORed with the current
// timestamp.
//
// See https://github.com/tdlib/td/blob/27d3fdd09d90f6b77ecbcce50b1e86dc4b3dd366/td/mtproto/TlsInit.cpp#L380-L384
// and https://tools.ietf.org/html/rfc5246#section-7.4.1.1.
func writeClientHello(
	w io.Writer,
	rand io.Reader,
	now clock.Clock,
	domain string,
	secret []byte,
) (r [32]byte, err error) {
	record, err := generateClientHello(rand, domain)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "generate ClientHello")
	}
	if len(record) < clientRandomOffset+clientRandomLength {
		return [32]byte{}, errors.Errorf("ClientHello is too short: %d bytes", len(record))
	}

	random := record[clientRandomOffset : clientRandomOffset+clientRandomLength]
	// Zero the ClientRandom before computing the digest, exactly as the proxy
	// does when it validates the handshake.
	for i := range random {
		random[i] = 0
	}

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(record); err != nil {
		return [32]byte{}, errors.Wrap(err, "hmac write")
	}
	copy(random, mac.Sum(nil))

	// Overwrite last 4 bytes using final := original ^ timestamp.
	old := binary.LittleEndian.Uint32(random[clientRandomLength-4:])
	old ^= uint32(now.Now().Unix())
	binary.LittleEndian.PutUint32(random[clientRandomLength-4:], old)

	// Copy ClientRandom for later use.
	copy(r[:], random)
	_, err = w.Write(record)
	return r, err
}
