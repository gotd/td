package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestRSAFingerprint(t *testing.T) {
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
	if RSAFingerprint(key) != 1914007313702140277 {
		t.Fatal(RSAFingerprint(key))
	}
}
