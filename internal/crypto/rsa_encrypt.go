package crypto

import (
	"crypto/rsa"
	"math/big"
)

// RSAEncrypt encrypts block with provided public key.
func RSAEncrypt(block [255]byte, key *rsa.PublicKey) []byte {
	z := big.NewInt(0).SetBytes(block[:])
	e := big.NewInt(int64(key.E))
	c := big.NewInt(0).Exp(z, e, key.N)

	res := make([]byte, 256)
	copy(res, c.Bytes())

	return res
}
