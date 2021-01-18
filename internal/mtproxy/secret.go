package mtproxy

import (
	"bytes"

	"golang.org/x/xerrors"
)

// SecretType represents MTProxy secret type.
type SecretType int

const (
	// Simple is a basic MTProxy secret.
	Simple SecretType = iota + 1
	// Secured is dd-secret.
	Secured
	// TLS is fakeTLS MTProxy secret.
	// First byte should be ee.
	TLS
)

// Secret represents MTProxy secret.
type Secret struct {
	DC        int
	Secret    []byte
	CloakHost string
	Type      SecretType
}

// ParseSecret checks and parses secret.
func ParseSecret(dc int, secret []byte) (Secret, error) {
	r := Secret{
		DC:     dc,
		Secret: secret,
	}
	const simpleLength = 16

	switch {
	case len(secret) == 1+simpleLength && bytes.HasPrefix(secret, []byte{0xdd}):
		r.Type = Secured

		secret = secret[1:]
		r.Secret = secret[:simpleLength]
	case len(secret) > simpleLength && bytes.HasPrefix(secret, []byte{0xee}):
		r.Type = TLS

		secret = secret[1:]
		r.Secret = secret[:simpleLength]
		r.CloakHost = string(secret[simpleLength:])
	case len(secret) == simpleLength:
		r.Type = Simple
	default:
		return Secret{}, xerrors.Errorf("invalid secret %q", string(secret))
	}

	return r, nil
}
