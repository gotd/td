package telegram

import (
	"crypto/rsa"

	"golang.org/x/xerrors"

	"github.com/ernado/td/internal/crypto"
	"github.com/ernado/td/telegram/internal"
)

//go:generate go run github.com/go-bindata/go-bindata/go-bindata -pkg=internal -o=internal/bindata.go -mode=420 -modtime=1 ./_data/...

// vendoredKeys parses vendored file _data/public_keys.pem as list of
// PEM-encoded public RSA keys.
//
// Most recent key list can be found on https://my.telegram.org/apps.
func vendoredKeys() ([]*rsa.PublicKey, error) {
	pem, err := internal.Asset("_data/public_keys.pem")
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	keys, err := crypto.ParseRSAPublicKeys(pem)
	if err != nil {
		return nil, xerrors.Errorf("parse: %w", err)
	}
	return keys, nil
}
