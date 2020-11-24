package crypto

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	mathrand "math/rand"
	"testing"
)

func TestRSAEncrypt(t *testing.T) {
	testKey := []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCJhOrkbrOi7/fFQRN2+8W5+Inx
bdxo7XH5DDKgFvXPDDe8cQINO3/kFat7KlpC2n2sS8ApZQkmREANg0XpImL9lCHB
v1FgQmL0xtnaURKo7FzaoaL4jCf5556NQr1th9F3oeN67mR4+BF0vPP9Gu6GY5Z1
BSqi+FEREW/2aWSgSwIDAQAB
-----END PUBLIC KEY-----`)
	b, _ := pem.Decode(testKey)
	keyParsed, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	key := keyParsed.(*rsa.PublicKey)
	rnd := mathrand.New(mathrand.NewSource(239))
	result, err := EncryptHashed([]byte("hello world"), key, rnd)
	if err != nil {
		t.Fatal(err)
	}
	expectedBase64 := "fxiV3AFhhauM5Dfgo50VzYBS9TxTMWPymU+cKt1HBwg9wPtIcXj3B2csCRSdCSCcjpBWJ" +
		"6NLk2QFsv+7IEeVHKXekpVBQ1aU4p52jPYSlqQZs0/BzYKxnexwD6qjYqvXJi60LD4S3fl7eCQnVoiL25vh" +
		"64F3a2cdYoiFQXgx7t4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAA=="
	gotBase64 := base64.StdEncoding.EncodeToString(result)
	if expectedBase64 != gotBase64 {
		t.Error(gotBase64)
	}
	if _, err := EncryptHashed(bytes.Repeat([]byte{1, 2, 3}, 1000), key, rnd); err == nil {
		t.Error("should error")
	}
}
