package mtproto

import (
	"crypto/rsa"
	// For embedding public keys.
	_ "embed"
	"sync"

	"github.com/gotd/td/internal/crypto"
)

var (
	//go:embed _data/public_keys.pem
	publicKeys []byte

	parsedKeys struct {
		Keys []*rsa.PublicKey
		Once sync.Once
	}
)

func init() {
	parsedKeys.Once.Do(func() {
		keys, err := crypto.ParseRSAPublicKeys(publicKeys)
		if err != nil {
			panic(err)
		}
		parsedKeys.Keys = keys
	})
}

// vendoredKeys parses vendored file _data/public_keys.pem as list of
// PEM-encoded public RSA keys.
//
// Most recent key list can be found on https://my.telegram.org/apps.
func vendoredKeys() []*rsa.PublicKey {
	return parsedKeys.Keys
}
