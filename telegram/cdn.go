package telegram

import (
	"crypto/rsa"
	"encoding/pem"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
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
			return nil, errors.Wrap(err, "parse RSA from PEM")
		}

		r = append(r, key)
	}

	return r, nil
}
