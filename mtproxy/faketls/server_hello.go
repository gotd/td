package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"github.com/go-faster/errors"
)

// readServerHello reads faketls ServerHello.
func readServerHello(r io.Reader, clientRandom [32]byte, secret []byte) error {
	packetBuf := bytes.NewBuffer(nil)
	r = io.TeeReader(r, packetBuf)

	handshake, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "handshake record")
	}
	if handshake.Type != RecordTypeHandshake {
		return errors.Wrap(err, "unexpected record type")
	}

	changeCipher, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "change cipher record")
	}
	if changeCipher.Type != RecordTypeChangeCipherSpec {
		return errors.Wrap(err, "unexpected record type")
	}

	cert, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "cert record")
	}
	if cert.Type != RecordTypeApplication {
		return errors.Wrap(err, "unexpected record type")
	}

	// `$record_header = type 1 byte + version 2 bytes + payload_length 2 bytes = 5 bytes`
	// `$server_hello_header = type 1 bytes + version 2 bytes + length 3 bytes = 6 bytes`
	// `$offset = $record_header + $server_hello_header = 11 bytes`
	const serverRandomOffset = 11
	packet := packetBuf.Bytes()
	// Copy original digest.
	var originalDigest [32]byte
	copy(originalDigest[:], packet[serverRandomOffset:serverRandomOffset+32])
	// Fill original digest by zeros.
	var zeros [32]byte
	copy(packet[serverRandomOffset:serverRandomOffset+32], zeros[:])

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
