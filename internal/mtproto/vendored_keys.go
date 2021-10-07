package mtproto

import (
	// For embedding public keys.
	_ "embed"
	"sync"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/exchange"
)

var (
	//go:embed _data/public_keys.pem
	publicKeys []byte

	// publicKeysNew is byte blob of new keys added for PQInnerData encryption (key exchange).
	//
	// See https://github.com/telegramdesktop/tdesktop/commit/95a7ce4622dc24717dc5b95fc99599dddfd4ff6c.
	//
	// See https://github.com/tdlib/td/commit/e9e24282378fcdb3a3ce020bee4253b65ac98213.
	//go:embed _data/public_keys_new.pem
	publicKeysNew []byte

	parsedKeys struct {
		Keys []exchange.PublicKey
		Once sync.Once
	}
)

//nolint:gochecknoinits
func init() {
	makePublicKeys := func(data []byte, rsaPad bool) ([]exchange.PublicKey, error) {
		rsaKeys, err := crypto.ParseRSAPublicKeys(data)
		if err != nil {
			return nil, err
		}

		keys := make([]exchange.PublicKey, 0, len(rsaKeys))
		for _, key := range rsaKeys {
			keys = append(keys, exchange.PublicKey{
				RSA:       key,
				UseRSAPad: rsaPad,
			})
		}
		return keys, nil
	}
	parsedKeys.Once.Do(func() {
		newKeys, err := makePublicKeys(publicKeysNew, true)
		if err != nil {
			panic(err)
		}
		parsedKeys.Keys = append(parsedKeys.Keys, newKeys...)

		oldKeys, err := makePublicKeys(publicKeys, false)
		if err != nil {
			panic(err)
		}
		parsedKeys.Keys = append(parsedKeys.Keys, oldKeys...)
	})
}

// vendoredKeys parses vendored file _data/public_keys.pem as list of
// PEM-encoded public RSA keys.
//
// Most recent key list can be found on https://my.telegram.org/apps.
func vendoredKeys() []exchange.PublicKey {
	return parsedKeys.Keys
}
