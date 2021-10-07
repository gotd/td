package telegram

import (
	"crypto/rsa"
	"encoding/pem"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
)

func parseCDNKeys(keys ...tg.CDNPublicKey) ([]*rsa.PublicKey, error) {
	r := make([]*rsa.PublicKey, 0, len(keys))

	for _, key := range keys {
		block, _ := pem.Decode([]byte(key.PublicKey))
		if block == nil {
			continue
		}

		key, err := crypto.ParseRSA(block.Bytes)
		if err != nil {
			return nil, xerrors.Errorf("parse RSA from PEM: %w", err)
		}

		r = append(r, key)
	}

	return r, nil
}
