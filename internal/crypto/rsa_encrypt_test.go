package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
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
	block := [255]byte{}
	rnd := mathrand.New(mathrand.NewSource(239))
	if _, err := io.ReadFull(rnd, block[:]); err != nil {
		t.Fatal(err)
	}
	copy(block[:], "hello world")
	result := RSAEncrypt(block, key)
	expectedBase64 := "GdYr3j/iIAYlM5Mn8Qzbpvr3oVcQ2Q0eUSBwAm1gkI2r3jQSy7Zg2a7FhOktDDh+a+A1rsVc+degM6a+d454XOzVaTDpK" +
		"Qdp8odBToE6nvhmux0fhCrrexLjWnoIjdl759Mf+bd2v0Db15LNoII8uI73Cnv8dQXCLgd4Mjqf/fAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
	if expectedBase64 != base64.StdEncoding.EncodeToString(result) {
		t.Error("encryption mismatch")
	}
}
