package crypto

import (
	"crypto/rsa"
	// #nosec
	//
	// Allowing sha1 because it is used in MTProto itself.
	"crypto/sha1"
	"math/big"
)

// RSAKeyBits is RSA key size.
//
// Can be used as rsa.GenerateKey(src, RSAKeyBits).
const RSAKeyBits = 2048

const (
	rsaLen         = 256
	rsaWithHashLen = 255
	rsaDataLen     = rsaWithHashLen - sha1.Size
)

// RSAPublicDecrypt recovers the message digest from the raw signature
// using the signer’s RSA public key.
//
// See also OpenSSL’s RSA_public_decrypt with RSA_NO_PADDING.
func RSAPublicDecrypt(pub *rsa.PublicKey, sig []byte) ([]byte, error) {
	k := pub.Size()
	if k < 11 || k != len(sig) {
		return nil, rsa.ErrVerification
	}

	c := new(big.Int).SetBytes(sig)
	e := big.NewInt(int64(pub.E))
	m := new(big.Int).Exp(c, e, pub.N)

	return m.Bytes(), nil
}

func rsaEncrypt(data []byte, key *rsa.PublicKey) []byte {
	z := new(big.Int).SetBytes(data)
	e := big.NewInt(int64(key.E))
	c := new(big.Int).Exp(z, e, key.N)
	res := make([]byte, rsaLen)
	c.FillBytes(res)
	return res
}

func rsaDecrypt(data []byte, key *rsa.PrivateKey, to []byte) bool {
	c := new(big.Int).SetBytes(data)
	m := new(big.Int).Exp(c, key.D, key.N)
	return FillBytes(m, to)
}
