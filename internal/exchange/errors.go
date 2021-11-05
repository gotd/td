package exchange

import (
	"github.com/ogen-go/errors"
)

// ErrKeyFingerprintNotFound is returned when client can't find keys by fingerprints
// provided by server during key exchange.
var ErrKeyFingerprintNotFound = errors.New("key fingerprint not found")
