package obfuscated2

import (
	"bytes"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type OneChar struct {
	char byte
}

func (c OneChar) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = c.char
	}
	return len(p), nil
}

func Test_generateKeys(t *testing.T) {
	a := require.New(t)
	secret, err := hex.DecodeString(strings.Repeat("a", 32))
	a.NoError(err)

	k, err := generateKeys(OneChar{char: 'a'}, [4]byte{0xdd, 0xdd, 0xdd, 0xdd}, secret, 2)
	a.NoError(err)

	var expectedHeader = []byte{
		97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97,
		97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97,
		97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97,
		97, 97, 97, 97, 97, 171, 65, 98, 66, 79, 102, 253, 220,
	}
	a.Equal(expectedHeader, k.header)
}

func TestEncrypt(t *testing.T) {
	a := require.New(t)
	secret, err := hex.DecodeString("8a96ef6e42a18c21837580cd1c91c5a8")
	a.NoError(err)

	rand := []byte{
		245, 118, 143, 80, 183, 49, 38, 10, 70, 190, 16, 39, 194, 238, 170,
		57, 53, 6, 36, 240, 182, 218, 89, 235, 165, 108, 129, 254, 69, 16,
		194, 224, 182, 29, 61, 211, 35, 238, 2, 56, 134, 51, 227, 131, 122,
		12, 28, 36, 250, 111, 41, 204, 215, 36, 190, 111, 65, 111, 247, 176,
		38, 246, 204, 230,
	}
	k, err := generateKeys(bytes.NewReader(rand), [4]byte{0xdd, 0xdd, 0xdd, 0xdd}, secret, 2)
	a.NoError(err)

	var expectedHeader = []byte{
		245, 118, 143, 80, 183, 49, 38, 10, 70, 190, 16, 39, 194, 238, 170, 57, 53, 6, 36, 240,
		182, 218, 89, 235, 165, 108, 129, 254, 69, 16, 194, 224, 182, 29, 61, 211, 35, 238,
		2, 56, 134, 51, 227, 131, 122, 12, 28, 36, 250, 111, 41, 204, 215, 36, 190, 111,
		190, 162, 221, 225, 109, 197, 157, 210,
	}
	a.Equal(expectedHeader, k.header)

	var encrypted [4]byte
	payload := []byte{'a', 'b', 'c', 'd'}
	k.encrypt.XORKeyStream(encrypted[:], payload)
	a.Equal([]byte{202, 122, 130, 38}, encrypted[:])

	k.decrypt.XORKeyStream(encrypted[:], payload)
	a.Equal([]byte{143, 113, 25, 130}, encrypted[:])
}

// chunkedReadWriter wraps a ReadWriter to return at most chunkSize bytes per Read.
// This simulates FakeTLS behavior where reads return one TLS record at a time.
type chunkedReadWriter struct {
	buf       bytes.Buffer
	chunkSize int
}

func (c *chunkedReadWriter) Write(p []byte) (int, error) {
	return c.buf.Write(p)
}

func (c *chunkedReadWriter) Read(p []byte) (int, error) {
	if len(p) > c.chunkSize {
		p = p[:c.chunkSize]
	}
	return c.buf.Read(p)
}

// TestReadPartialChunks verifies that Obfuscated2 correctly decrypts data
// when the underlying reader returns fewer bytes than requested (as FakeTLS
// does when a large payload spans multiple TLS records).
//
// This is a regression test for a bug where XORKeyStream was called on the
// full buffer slice instead of only the bytes actually read, causing the
// AES-CTR stream to advance too far and corrupt subsequent decryption.
func TestReadPartialChunks(t *testing.T) {
	a := require.New(t)

	// Create matching AES-CTR streams for encrypt and decrypt.
	key := make([]byte, 32)
	iv := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	encStream, err := createCTR(key, iv)
	a.NoError(err)
	decStream, err := createCTR(key, iv)
	a.NoError(err)

	// Create a large payload (50KB) that will require multiple partial reads.
	const payloadSize = 50000
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i % 251) // deterministic pattern
	}

	// Encrypt the payload.
	encrypted := make([]byte, payloadSize)
	encStream.XORKeyStream(encrypted, payload)

	// Read through Obfuscated2 with a chunked underlying reader (16KB chunks).
	// This simulates FakeTLS returning one TLS record at a time.
	readBuf := &chunkedReadWriter{chunkSize: 16384}
	readBuf.buf.Write(encrypted)

	reader := &Obfuscated2{conn: readBuf, keys: keys{decrypt: decStream}}

	// Use io.ReadFull which passes large buffer slices to Read().
	// Before the fix, XORKeyStream(b, b) would advance the stream by len(b)
	// instead of n, corrupting all subsequent decryption.
	result := make([]byte, payloadSize)
	n, err := io.ReadFull(reader, result)
	a.NoError(err)
	a.Equal(payloadSize, n)
	a.Equal(payload, result, "decrypted data should match original payload")
}

func Test_getDecryptInit(t *testing.T) {
	a := require.New(t)
	var input [64]byte
	for i := range input {
		input[i] = byte(i)
	}

	r := getDecryptInit(input[:])
	expected := [48]byte{
		55, 54, 53, 52, 51, 50, 49, 48, 47, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 32,
		31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8,
	}
	a.Equal(expected, r)
}
