package exchange

import (
	"crypto/rsa"

	"github.com/gotd/td/crypto"
)

// PublicKey is a public Telegram server key.
type PublicKey struct {
	// RSA public key.
	RSA *rsa.PublicKey
}

// Zero denotes that current PublicKey is zero value.
func (k PublicKey) Zero() bool {
	return k.RSA == nil
}

// Fingerprint computes key fingerprint.
func (k PublicKey) Fingerprint() int64 {
	return crypto.RSAFingerprint(k.RSA)
}

// PrivateKey is a private Telegram server key.
type PrivateKey struct {
	// RSA private key.
	RSA *rsa.PrivateKey
}

// Zero denotes that current PublicKey is zero value.
func (k PrivateKey) Zero() bool {
	return k.RSA == nil
}

// Fingerprint computes key fingerprint.
func (k PrivateKey) Fingerprint() int64 {
	return crypto.RSAFingerprint(&k.RSA.PublicKey)
}

// Public returns PublicKey of this PrivateKey pair.
func (k PrivateKey) Public() PublicKey {
	return PublicKey{
		RSA: &k.RSA.PublicKey,
	}
}
