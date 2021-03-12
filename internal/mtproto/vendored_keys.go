package mtproto

import (
	"crypto/rsa"
	// For embedding public keys.
	_ "embed"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
)

//go:embed _data/public_keys.pem
var publicKeys []byte // nolint:gochecknoglobals

// vendoredKeys parses vendored file _data/public_keys.pem as list of
// PEM-encoded public RSA keys.
//
// Most recent key list can be found on https://my.telegram.org/apps.
func vendoredKeys() ([]*rsa.PublicKey, error) {
	keys, err := crypto.ParseRSAPublicKeys(publicKeys)
	if err != nil {
		return nil, xerrors.Errorf("parse: %w", err)
	}
	return keys, nil
}
