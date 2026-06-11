package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"
)

// checkClientHello replicates the acceptance checks performed by the MTProxy
// server on the FakeTLS ClientHello.
//
// See net-tcp-rpc-ext-server.c in the reference MTProxy implementation.
func checkClientHello(t *testing.T, record []byte, domain string, secret []byte, now time.Time) {
	t.Helper()
	a := require.New(t)

	// Well-formed TLS handshake record.
	a.Greater(len(record), 5, "record header")
	a.Equal(byte(RecordTypeHandshake), record[0], "record type")
	recordLen := int(binary.BigEndian.Uint16(record[3:5]))
	a.Equal(len(record)-5, recordLen, "record length matches body")
	a.Equal(byte(0x01), record[5], "handshake type ClientHello")

	// SNI must carry the cloak domain.
	a.True(bytes.Contains(record, []byte(domain)), "domain present in ClientHello")

	// Session ID must be 32 bytes long: the server echoes client_hello+44 as a
	// 32-byte session ID, so offset 43 must hold the length byte 0x20.
	a.Equal(byte(0x20), record[43], "session ID length")

	// Cipher suites: after skipping GREASE, the first suite must be a TLS 1.3
	// suite (0x13 0x01..0x03).
	pos := 76
	cipherSuitesLen := int(binary.BigEndian.Uint16(record[pos:]))
	pos += 2
	a.LessOrEqual(pos+cipherSuitesLen, len(record), "cipher suites length")
	for cipherSuitesLen >= 2 && record[pos]&0x0F == 0x0A && record[pos+1]&0x0F == 0x0A {
		cipherSuitesLen -= 2
		pos += 2
	}
	a.Greater(cipherSuitesLen, 1, "supported cipher suite present")
	a.Equal(byte(0x13), record[pos], "TLS 1.3 cipher suite")
	a.GreaterOrEqual(record[pos+1], byte(0x01))
	a.LessOrEqual(record[pos+1], byte(0x03))

	// Digest: HMAC-SHA256 over the record with ClientRandom zeroed must match
	// the first 28 bytes of ClientRandom.
	var clientRandom [32]byte
	copy(clientRandom[:], record[clientRandomOffset:clientRandomOffset+clientRandomLength])

	zeroed := append([]byte(nil), record...)
	for i := clientRandomOffset; i < clientRandomOffset+clientRandomLength; i++ {
		zeroed[i] = 0
	}
	mac := hmac.New(sha256.New, secret)
	_, err := mac.Write(zeroed)
	a.NoError(err)
	expected := mac.Sum(nil)
	a.Equal(expected[:28], clientRandom[:28], "digest matches")

	// Timestamp is the last 4 bytes of the digest XORed with ClientRandom.
	timestamp := binary.LittleEndian.Uint32(expected[28:]) ^ binary.LittleEndian.Uint32(clientRandom[28:])
	a.Equal(uint32(now.Unix()), timestamp, "timestamp")
}

func TestTLS(t *testing.T) {
	a := require.New(t)
	secret := []byte("0123456789abcdef")
	const domain = "google.com"
	now := time.Date(2010, 10, 10, 1, 1, 1, 0, time.UTC)
	c := neo.NewTime(now)

	b := bytes.NewBuffer(nil)
	digest, err := writeClientHello(b, rand.Reader, c, domain, secret)
	a.NoError(err)

	record := b.Bytes()
	checkClientHello(t, record, domain, secret, now)
	// Returned digest is the ClientRandom that the server hello is verified against.
	a.Equal(record[clientRandomOffset:clientRandomOffset+clientRandomLength], digest[:])

	// ClientHello must be randomized between calls.
	b2 := bytes.NewBuffer(nil)
	_, err = writeClientHello(b2, rand.Reader, c, domain, secret)
	a.NoError(err)
	a.NotEqual(record, b2.Bytes(), "ClientHello must be randomized")
}

func TestFakeTLSRead_SkipsChangeCipherSpecRecords(t *testing.T) {
	a := require.New(t)

	conn := bytes.NewBuffer(nil)
	_, err := writeRecord(conn, record{
		Type:    RecordTypeChangeCipherSpec,
		Version: Version12Bytes,
		Data:    []byte{0x01},
	})
	a.NoError(err)
	_, err = writeRecord(conn, record{
		Type:    RecordTypeApplication,
		Version: Version12Bytes,
		Data:    []byte("hello"),
	})
	a.NoError(err)

	f := NewFakeTLS(bytes.NewReader(nil), conn)
	got := make([]byte, len("hello"))
	n, err := io.ReadFull(f, got)
	a.NoError(err)
	a.Equal(len("hello"), n)
	a.Equal("hello", string(got))
}

func TestFakeTLSRead_RejectsUnexpectedRecordTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		rec  record
		msg  string
	}{
		{
			name: "handshake",
			rec: record{
				Type:    RecordTypeHandshake,
				Version: Version12Bytes,
				Data:    []byte{0x01},
			},
			msg: "unexpected record type handshake",
		},
		{
			name: "unsupported",
			rec: record{
				Type:    RecordTypeAlert,
				Version: Version12Bytes,
				Data:    []byte{0x01},
			},
			msg: "unsupported record type",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := require.New(t)

			conn := bytes.NewBuffer(nil)
			_, err := writeRecord(conn, tc.rec)
			a.NoError(err)

			f := NewFakeTLS(bytes.NewReader(nil), conn)
			buf := make([]byte, 1)
			_, err = f.Read(buf)
			a.Error(err)
			a.Contains(err.Error(), tc.msg)
		})
	}
}
