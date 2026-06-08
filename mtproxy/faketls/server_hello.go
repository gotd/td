package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"github.com/go-faster/errors"
)

// Sanity limit for the number of extra handshake records before
// ChangeCipherSpec.
const maxHandshakeRecords = 16

// readServerHello reads faketls ServerHello.
func readServerHello(r io.Reader, clientRandom [32]byte, secret []byte) error {
	packetBuf := bytes.NewBuffer(nil)
	r = io.TeeReader(r, packetBuf)

	handshake, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "handshake record")
	}
	if handshake.Type != RecordTypeHandshake {
		return errors.New("unexpected record type")
	}

	// `$record_header = type 1 byte + version 2 bytes + payload_length 2 bytes = 5 bytes`
	// `$server_hello_header = type 1 bytes + version 2 bytes + length 3 bytes = 6 bytes`
	// `$offset = $record_header + $server_hello_header = 11 bytes`
	const serverRandomOffset = 11
	const serverRandomEnd = serverRandomOffset + 32
	if packetBuf.Len() < serverRandomEnd {
		return errors.New("handshake record is too short")
	}

	changeCipherFound := false
	for i := 0; i < maxHandshakeRecords; i++ {
		rec, err := readRecord(r)
		if err != nil {
			return errors.Wrap(err, "change cipher record")
		}

		switch rec.Type {
		case RecordTypeHandshake:
			continue
		case RecordTypeChangeCipherSpec:
			changeCipherFound = true
		default:
			return errors.New("unexpected record type")
		}

		break
	}

	if !changeCipherFound {
		return errors.New("unexpected record type")
	}

	cert, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "cert record")
	}
	if cert.Type != RecordTypeApplication {
		return errors.New("unexpected record type")
	}

	packet := packetBuf.Bytes()
	var originalDigest [32]byte
	copy(originalDigest[:], packet[serverRandomOffset:serverRandomEnd])
	// Fill original digest by zeros.
	var zeros [32]byte
	copy(packet[serverRandomOffset:serverRandomEnd], zeros[:])

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(clientRandom[:]); err != nil {
		return errors.Wrap(err, "hmac write")
	}
	if _, err := mac.Write(packet); err != nil {
		return errors.Wrap(err, "hmac write")
	}
	if !bytes.Equal(mac.Sum(nil), originalDigest[:]) {
		return errors.New("hmac digest mismatch")
	}

	return nil
}
