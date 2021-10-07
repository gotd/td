package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/testutil"
)

func TestRSAPad(t *testing.T) {
	a := require.New(t)

	keys, err := ParseRSAPublicKeys([]byte(`
-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA6LszBcC1LGzyr992NzE0ieY+BSaOW622Aa9Bd4ZHLl+TuFQ4lo4g
5nKaMBwK/BIb9xUfg0Q29/2mgIR6Zr9krM7HjuIcCzFvDtr+L0GQjae9H0pRB2OO
62cECs5HKhT5DZ98K33vmWiLowc621dQuwKWSQKjWf50XYFw42h21P2KXUGyp2y/
+aEyZ+uVgLLQbRA1dEjSDZ2iGRy12Mk5gpYc397aYp438fsJoHIgJ2lgMv5h7WY9
t6N/byY9Nw9p21Og3AoXSL2q/2IJ1WRUhebgAdGVMlV1fkuOQoEzR7EdpqtQD9Cs
5+bfo3Nhmcyvk5ftB0WkJ9z6bNZ7yxrP8wIDAQAB
-----END RSA PUBLIC KEY-----`))
	a.NoError(err)
	data := bytes.Repeat([]byte{'a'}, 144)

	encrypted, err := RSAPad(data, keys[0], Zero{})
	a.NoError(err)
	a.Len(encrypted, 256)

	hexResult := "bf68719e836806b040cd261ecaf66eb3c4ba19f3bbea3031b2e6cf29167bab647201d101b291dc" +
		"5b716a42e789a38d947fe59e9bcce8f30ef46a946743ea8b6babbce7fc0afc46b802aa453e83471d82a4dfad83f971f35" +
		"0b4b4fb474cd1c48fdf427e4b5fecce9ec3178ae7dac3985856fdefa21d6fdc5e0e0fd8a57bc4f51580d637d372be8d87" +
		"c9aa3fde8e6f8287bcb3be846aadcdd59465375479e248f62ed438f9804fbe36d41ca906243a5f740f3937949aa149ba8" +
		"a8b8e68b3f3e1e3cd3f946387520e21eee55845e1f015a919a22f6a72bfaecd2cae946c91983b41f9ffabe97963bbde8f" +
		"30eaf5fd3c5b8cecab8711bd269e441b6084f385726ff0"
	expected, err := hex.DecodeString(hexResult)
	a.NoError(err)
	a.Equal(expected, encrypted)
}

func TestDecodeRSAPad(t *testing.T) {
	a := require.New(t)
	r := rand.Reader

	key, err := rsa.GenerateKey(r, RSAKeyBits)
	a.NoError(err)
	size := 144

	data := make([]byte, size)
	_, err = io.ReadFull(r, data)
	a.NoError(err)

	encrypted, err := RSAPad(data, &key.PublicKey, r)
	a.NoError(err)
	a.Len(encrypted, 256)

	decrypted, err := DecodeRSAPad(encrypted, key)
	a.NoError(err)
	a.Equal(data, decrypted[:size])
}

func BenchmarkRSAPad(b *testing.B) {
	key := testutil.RSAPrivateKey()

	data := make([]byte, 144)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := RSAPad(data, &key.PublicKey, rand.Reader); err != nil {
			b.Fatal(err)
		}
	}
}
