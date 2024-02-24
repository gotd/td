package mtproxy

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/proto/codec"
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
	Secret    []byte
	Tag       byte
	CloakHost string
	Type      SecretType
}

// ExpectedCodec returns codec from secret tag if it exists.
func (s Secret) ExpectedCodec() (cdc codec.Codec, _ bool) {
	switch s.Tag {
	case codec.AbridgedClientStart[0]:
		cdc = codec.Abridged{}
	case codec.IntermediateClientStart[0]:
		cdc = codec.Intermediate{}
	case codec.PaddedIntermediateClientStart[0]:
		cdc = codec.PaddedIntermediate{}
	default:
		return nil, false
	}

	return cdc, true
}

// ParseSecret checks and parses secret.
func ParseSecret(secret []byte) (Secret, error) {
	r := Secret{
		Secret: secret,
	}
	const simpleLength = 16

	switch {
	case len(secret) == 1+simpleLength:
		r.Type = Secured

		r.Tag = secret[0]
		secret = secret[1:]
		r.Secret = secret[:simpleLength]
	case len(secret) > simpleLength:
		r.Type = TLS

		r.Tag = secret[0]
		secret = secret[1:]
		r.Secret = secret[:simpleLength]
		r.CloakHost = string(secret[simpleLength:])
	case len(secret) == simpleLength:
		r.Type = Simple
	default:
		return Secret{}, errors.Errorf("invalid secret %q", string(secret))
	}

	if r.Type != Simple {
		switch r.Tag {
		case codec.AbridgedClientStart[0],
			codec.IntermediateClientStart[0],
			codec.PaddedIntermediateClientStart[0]:
		default:
			return Secret{}, errors.Errorf("unknown tag %+x", r.Tag)
		}
	}

	return r, nil
}
