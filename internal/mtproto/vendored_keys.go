package mtproto

import (
	// For embedding public keys.
	_ "embed"
	"sync"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/exchange"
)

var (
	//go:embed _data/public_keys.pem
	publicKeys []byte

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
