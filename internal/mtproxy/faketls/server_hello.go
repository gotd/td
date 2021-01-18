package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"golang.org/x/xerrors"
)

// readServerHello reads faketls ServerHello.
// See https://github.com/9seconds/mtg/blob/e075169dd4e9fc4c2b1453668f85f5099c4fb895/tlstypes/server_hello.go#L21-L57.
func readServerHello(r io.Reader, clientRandom [32]byte, secret []byte) error {
	packetBuf := bytes.NewBuffer(nil)
	r = io.TeeReader(r, packetBuf)

	handshake, err := readRecord(r)
	if err != nil {
		return xerrors.Errorf("handshake record: %w", err)
	}
	if handshake.Type != RecordTypeHandshake {
		return xerrors.Errorf("unexpected record type: %w", err)
	}

	changeCipher, err := readRecord(r)
	if err != nil {
		return xerrors.Errorf("change cipher record: %w", err)
	}
	if changeCipher.Type != RecordTypeChangeCipherSpec {
		return xerrors.Errorf("unexpected record type: %w", err)
	}

	cert, err := readRecord(r)
	if err != nil {
		return xerrors.Errorf("cert record: %w", err)
	}
	if cert.Type != RecordTypeApplication {
		return xerrors.Errorf("unexpected record type: %w", err)
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
		return xerrors.Errorf("hmac write: %w", err)
	}
	if _, err := mac.Write(packet); err != nil {
		return xerrors.Errorf("hmac write: %w", err)
	}
	if !bytes.Equal(mac.Sum(nil), originalDigest[:]) {
		return xerrors.New("hmac digest mismatch")
	}

	return nil
}
