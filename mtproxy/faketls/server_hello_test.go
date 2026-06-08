package faketls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_readServerHello(t *testing.T) {
	p := filepath.Join("testdata", "server_hello")
	entries, err := os.ReadDir(p)
	if err != nil {
		t.Fatal(err)
	}

	clientRandom := map[string][32]uint8{
		"alexbers.hex": {
			0xa1, 0x32, 0xe3, 0x91, 0x60, 0x83, 0xb3, 0x14, 0xc1, 0xb9, 0x74, 0xd0, 0x57, 0x85, 0xe8, 0xee,
			0x70, 0x45, 0x6e, 0x5f, 0x86, 0x6d, 0x96, 0x57, 0xd5, 0x0a, 0x5c, 0x08, 0xea, 0x38, 0x31, 0x8e,
		},
		"mtgv2.hex": {
			0xca, 0x36, 0xbb, 0xa8, 0x33, 0x80, 0x9a, 0x33, 0xaa, 0x62, 0x7e, 0xbb, 0x5a, 0x32, 0xa1, 0x01,
			0x02, 0xd1, 0xa6, 0x1e, 0x1e, 0x6c, 0x58, 0xa4, 0x61, 0xd9, 0x34, 0x57, 0x4d, 0x2e, 0x2e, 0xa3,
		},
	}
	secret := []uint8{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Error("Unexpected directory in testdata")
			continue
		}
		fileName := entry.Name()

		t.Run(fileName, func(t *testing.T) {
			a := require.New(t)

			data, err := os.ReadFile(filepath.Join(p, fileName))
			a.NoError(err)

			decode, err := hex.DecodeString(strings.TrimSpace(string(data)))
			a.NoError(err)

			r := bytes.NewReader(decode)
			a.NoError(readServerHello(r, clientRandom[entry.Name()], secret))

			a.Zero(r.Len(), "should read all data")
		})
	}
}

func Test_readServerHello_AllowsAdditionalHandshakeRecords(t *testing.T) {
	a := require.New(t)

	clientRandom := [32]byte{
		0xa1, 0x32, 0xe3, 0x91, 0x60, 0x83, 0xb3, 0x14, 0xc1, 0xb9, 0x74, 0xd0, 0x57, 0x85, 0xe8, 0xee,
		0x70, 0x45, 0x6e, 0x5f, 0x86, 0x6d, 0x96, 0x57, 0xd5, 0x0a, 0x5c, 0x08, 0xea, 0x38, 0x31, 0x8e,
	}
	secret := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}

	packet := bytes.NewBuffer(nil)
	_, err := writeRecord(packet, record{
		Type:    RecordTypeHandshake,
		Version: Version12Bytes,
		Data:    make([]byte, 38),
	})
	a.NoError(err)
	_, err = writeRecord(packet, record{
		Type:    RecordTypeHandshake,
		Version: Version12Bytes,
		Data:    []byte{0x0b, 0x00, 0x00, 0x00},
	})
	a.NoError(err)
	_, err = writeRecord(packet, record{
		Type:    RecordTypeHandshake,
		Version: Version12Bytes,
		Data:    []byte{0x0c, 0x00, 0x00, 0x00},
	})
	a.NoError(err)
	_, err = writeRecord(packet, record{
		Type:    RecordTypeChangeCipherSpec,
		Version: Version12Bytes,
		Data:    []byte{0x01},
	})
	a.NoError(err)
	_, err = writeRecord(packet, record{
		Type:    RecordTypeApplication,
		Version: Version12Bytes,
		Data:    []byte{0x14, 0x00, 0x00, 0x00},
	})
	a.NoError(err)

	raw := packet.Bytes()
	const serverRandomOffset = 11
	const serverRandomEnd = serverRandomOffset + 32

	mac := hmac.New(sha256.New, secret)
	_, err = mac.Write(clientRandom[:])
	a.NoError(err)
	_, err = mac.Write(raw)
	a.NoError(err)
	copy(raw[serverRandomOffset:serverRandomEnd], mac.Sum(nil))

	a.NoError(readServerHello(bytes.NewReader(raw), clientRandom, secret))
}

func Test_readServerHello_ErrorPaths(t *testing.T) {
	clientRandom := [32]byte{}
	secret := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}

	for _, tc := range []struct {
		name string
		buf  func(t *testing.T) []byte
		msg  string
	}{
		{
			name: "first record is not handshake",
			buf: func(t *testing.T) []byte {
				t.Helper()
				b := bytes.NewBuffer(nil)
				_, err := writeRecord(b, record{Type: RecordTypeApplication, Version: Version12Bytes, Data: []byte{0x01}})
				require.NoError(t, err)
				return b.Bytes()
			},
			msg: "unexpected record type",
		},
		{
			name: "first handshake too short",
			buf: func(t *testing.T) []byte {
				t.Helper()
				b := bytes.NewBuffer(nil)
				_, err := writeRecord(b, record{Type: RecordTypeHandshake, Version: Version12Bytes, Data: []byte{0x01}})
				require.NoError(t, err)
				return b.Bytes()
			},
			msg: "handshake record is too short",
		},
		{
			name: "unexpected record before change cipher",
			buf: func(t *testing.T) []byte {
				t.Helper()
				b := bytes.NewBuffer(nil)
				_, err := writeRecord(b, record{Type: RecordTypeHandshake, Version: Version12Bytes, Data: make([]byte, 38)})
				require.NoError(t, err)
				_, err = writeRecord(b, record{Type: RecordTypeApplication, Version: Version12Bytes, Data: []byte{0x01}})
				require.NoError(t, err)
				return b.Bytes()
			},
			msg: "unexpected record type",
		},
		{
			name: "application after change cipher is missing",
			buf: func(t *testing.T) []byte {
				t.Helper()
				b := bytes.NewBuffer(nil)
				_, err := writeRecord(b, record{Type: RecordTypeHandshake, Version: Version12Bytes, Data: make([]byte, 38)})
				require.NoError(t, err)
				_, err = writeRecord(b, record{Type: RecordTypeChangeCipherSpec, Version: Version12Bytes, Data: []byte{0x01}})
				require.NoError(t, err)
				_, err = writeRecord(b, record{Type: RecordTypeHandshake, Version: Version12Bytes, Data: []byte{0x01}})
				require.NoError(t, err)
				return b.Bytes()
			},
			msg: "unexpected record type",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := require.New(t)
			err := readServerHello(bytes.NewReader(tc.buf(t)), clientRandom, secret)
			a.Error(err)
			a.Contains(err.Error(), tc.msg)
		})
	}
}
