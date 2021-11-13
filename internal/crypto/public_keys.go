package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/go-faster/errors"
)

// ParseRSAPublicKeys parses data as list of PEM-encdoed public keys.
func ParseRSAPublicKeys(data []byte) ([]*rsa.PublicKey, error) {
	var keys []*rsa.PublicKey

	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		key, err := ParseRSA(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "parse RSA from PEM")
		}

		keys = append(keys, key)
		data = rest
	}

	return keys, nil
}

// ParseRSA parses data RSA key in PKCS1 or PKIX forms.
func ParseRSA(data []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKCS1PublicKey(data)
	if err == nil {
		return key, nil
	}
	k, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	publicKey, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, errors.Errorf("parsed unexpected key type %T", k)
	}
	return publicKey, nil
}
