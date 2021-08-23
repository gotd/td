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

	parsedKeys struct {
		Keys []exchange.PublicKey
		Once sync.Once
	}
)

//nolint:gochecknoinits
func init() {
	parsedKeys.Once.Do(func() {
		rsaKeys, err := crypto.ParseRSAPublicKeys(publicKeys)
		if err != nil {
			panic(err)
		}

		keys := make([]exchange.PublicKey, 0, len(rsaKeys))
		for _, key := range rsaKeys {
			// TODO(tdakkota): distinguish new and old keys via UseInnerDataDC.
			keys = append(keys, exchange.PublicKey{
				RSA:            key,
				UseInnerDataDC: false,
			})
		}

		parsedKeys.Keys = keys
	})
}

// vendoredKeys parses vendored file _data/public_keys.pem as list of
// PEM-encoded public RSA keys.
//
// Most recent key list can be found on https://my.telegram.org/apps.
func vendoredKeys() []exchange.PublicKey {
	return parsedKeys.Keys
}
