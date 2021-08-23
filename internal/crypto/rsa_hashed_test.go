package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	mathrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRSAEncryptHashed(t *testing.T) {
	testKey := []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCJhOrkbrOi7/fFQRN2+8W5+Inx
bdxo7XH5DDKgFvXPDDe8cQINO3/kFat7KlpC2n2sS8ApZQkmREANg0XpImL9lCHB
v1FgQmL0xtnaURKo7FzaoaL4jCf5556NQr1th9F3oeN67mR4+BF0vPP9Gu6GY5Z1
BSqi+FEREW/2aWSgSwIDAQAB
-----END PUBLIC KEY-----`)
	keys, err := ParseRSAPublicKeys(testKey)
	if err != nil {
		t.Fatal(err)
	}
	key := keys[0]
	rnd := mathrand.New(mathrand.NewSource(239))
	result, err := RSAEncryptHashed([]byte("hello world"), key, rnd)
	if err != nil {
		t.Fatal(err)
	}
	expectedBase64 := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB/GJXcAWGFq4zkN+CjnRXNgFL1" +
		"PFMxY/KZT5wq3UcHCD3A+0hxePcHZywJFJ0JIJyOkFYno0uTZAWy/7sgR5Ucpd6SlUFDVp" +
		"TinnaM9hKWpBmzT8HNgrGd7HAPqqNiq9cmLrQsPhLd+Xt4JCdWiIvbm+HrgXdrZx1iiIVB" +
		"eDHu3g=="
	gotBase64 := base64.StdEncoding.EncodeToString(result)
	if expectedBase64 != gotBase64 {
		t.Error(gotBase64)
	}
	if _, err := RSAEncryptHashed(bytes.Repeat([]byte{1, 2, 3}, 1000), key, rnd); err == nil {
		t.Error("should error")
	}
}

func TestRSADecryptHashed(t *testing.T) {
	a := require.New(t)
	src := rand.Reader
	k, err := rsa.GenerateKey(src, RSAKeyBits)
	a.NoError(err)

	plaintext := []byte("abcd")
	encrypted, err := RSAEncryptHashed(plaintext, &k.PublicKey, src)
	a.NoError(err)
	decrypted, err := RSADecryptHashed(encrypted, k)
	a.NoError(err)
	a.Equal(plaintext, decrypted)
}

func TestRSAEncryptHashedCorpus(t *testing.T) {
	reader := mathrand.New(mathrand.NewSource(0))
	k, err := rsa.GenerateKey(reader, RSAKeyBits)
	require.NoError(t, err)

	for _, s := range []string{
		"\xbd\xbf\xef\x1e\x11p",
	} {
		t.Run(s, func(t *testing.T) {
			data := []byte(s)
			encrypted, err := RSAEncryptHashed(data, &k.PublicKey, reader)
			require.NoError(t, err)

			decrypted, err := RSADecryptHashed(encrypted, k)
			require.NoError(t, err)

			require.Equal(t, data, decrypted)
		})
	}
}

func TestRSAEncryptHashedFuzz(t *testing.T) {
	src := mathrand.New(mathrand.NewSource(1))
	k, err := rsa.GenerateKey(src, RSAKeyBits)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		n, err := RandInt64n(src, rsaDataLen)
		require.NoError(t, err)
		data := make([]byte, int(n))
		encrypted, err := RSAEncryptHashed(data, &k.PublicKey, src)
		require.NoError(t, err)
		decrypted, err := RSADecryptHashed(encrypted, k)
		require.NoError(t, err)
		require.Equal(t, data, decrypted)
	}
}
