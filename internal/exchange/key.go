package exchange

import (
	"crypto/rsa"

	"github.com/nnqq/td/internal/crypto"
)

// PublicKey is a public Telegram server key.
type PublicKey struct {
	// RSA public key.
	RSA *rsa.PublicKey
	// UseRSAPad denotes to use crypto.RSAPad instead of crypto.RSAEncryptHashed.
	//
	// See https://github.com/tdlib/td/commit/e9e24282378fcdb3a3ce020bee4253b65ac98213.
	UseRSAPad bool
}

// Zero denotes that current PublicKey is zero value.
func (k PublicKey) Zero() bool {
	return k.RSA == nil && !k.UseRSAPad
}

// Fingerprint computes key fingerprint.
func (k PublicKey) Fingerprint() int64 {
	return crypto.RSAFingerprint(k.RSA)
}

// PrivateKey is a private Telegram server key.
type PrivateKey struct {
	// RSA private key.
	RSA *rsa.PrivateKey
	// UseRSAPad denotes to use crypto.RSAPad instead of crypto.RSAEncryptHashed.
	//
	// See https://github.com/tdlib/td/commit/e9e24282378fcdb3a3ce020bee4253b65ac98213.
	UseRSAPad bool
}

// Zero denotes that current PublicKey is zero value.
func (k PrivateKey) Zero() bool {
	return k.RSA == nil && !k.UseRSAPad
}

// Fingerprint computes key fingerprint.
func (k PrivateKey) Fingerprint() int64 {
	return crypto.RSAFingerprint(&k.RSA.PublicKey)
}

// Public returns PublicKey of this PrivateKey pair.
func (k PrivateKey) Public() PublicKey {
	return PublicKey{
		RSA:       &k.RSA.PublicKey,
		UseRSAPad: k.UseRSAPad,
	}
}
