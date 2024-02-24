package obfuscated2

import (
	"bytes"
	"encoding/hex"
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
