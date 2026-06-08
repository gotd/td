package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"github.com/go-faster/errors"
)

const maxHandshakeRecords = 16

// readServerHello reads faketls ServerHello.
func readServerHello(r io.Reader, clientRandom [32]byte, secret []byte) error {
	packetBuf := bytes.NewBuffer(nil)
	r = io.TeeReader(r, packetBuf)

	firstRec, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "first handshake record")
	}
	if firstRec.Type != RecordTypeHandshake {
		return errors.Errorf("expected Handshake record first, got type 0x%02x", firstRec.Type)
	}

	const serverRandomOffset = 11
	const serverRandomEnd = serverRandomOffset + 32
	if packetBuf.Len() < serverRandomEnd {
		return errors.Errorf("first Handshake record too short: %d bytes, need at least %d", packetBuf.Len(), serverRandomEnd)
	}

	changeCipherFound := false
	for i := 1; i <= maxHandshakeRecords; i++ {
		rec, err := readRecord(r)
		if err != nil {
			return errors.Wrapf(err, "record[%d]", i)
		}

		switch rec.Type {
		case RecordTypeHandshake:
			continue
		case RecordTypeChangeCipherSpec:
			changeCipherFound = true
		default:
			return errors.Errorf("unexpected record type 0x%02x before ChangeCipherSpec", rec.Type)
		}

		if changeCipherFound {
			break
		}
	}

	if !changeCipherFound {
		return errors.Errorf("ChangeCipherSpec not found within %d records", maxHandshakeRecords)
	}

	appRec, err := readRecord(r)
	if err != nil {
		return errors.Wrap(err, "application record after ChangeCipherSpec")
	}
	if appRec.Type != RecordTypeApplication {
		return errors.Errorf("expected Application record after ChangeCipherSpec, got type 0x%02x", appRec.Type)
	}

	packet := packetBuf.Bytes()
	var originalDigest [32]byte
	copy(originalDigest[:], packet[serverRandomOffset:serverRandomEnd])

	var zeros [32]byte
	copy(packet[serverRandomOffset:serverRandomEnd], zeros[:])

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(clientRandom[:]); err != nil {
		return errors.Wrap(err, "hmac write clientRandom")
	}
	if _, err := mac.Write(packet); err != nil {
		return errors.Wrap(err, "hmac write packet")
	}
	if !bytes.Equal(mac.Sum(nil), originalDigest[:]) {
		return errors.New("hmac digest mismatch")
	}

	return nil
}
