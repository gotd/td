package crypto

import (
	"bytes"
	"crypto/rsa"

	// #nosec
	//
	// Allowing sha1 because it is used in MTProto itself.
	"crypto/sha1"
	"io"
	"math/big"

	"golang.org/x/xerrors"
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

// RSAEncryptHashed encrypts given data with RSA, prefixing with a hash.
func RSAEncryptHashed(data []byte, key *rsa.PublicKey, randomSource io.Reader) ([]byte, error) {
	// Preparing `data_with_hash`.
	// data_with_hash := SHA1(data) + data + (any random bytes);
	// such that the length equals 255 bytes;
	var dataWithHash [rsaWithHashLen]byte
	if len(data) > rsaDataLen {
		return nil, xerrors.Errorf("data length %d is too big", len(data))
	}

	// Filling data_with_hash with random bytes.
	if _, err := io.ReadFull(randomSource, dataWithHash[:]); err != nil {
		return nil, err
	}

	h := sha1.Sum(data) // #nosec

	// Replacing first 20 bytes with sha1(data).
	copy(dataWithHash[:sha1.Size], h[:])
	// Replacing other bytes with data itself.
	copy(dataWithHash[sha1.Size:], data)

	// Encrypting "dataWithHash" with RSA.
	z := new(big.Int).SetBytes(dataWithHash[:])
	e := big.NewInt(int64(key.E))
	c := new(big.Int).Exp(z, e, key.N)
	res := make([]byte, rsaLen)
	c.FillBytes(res)

	return res, nil
}

// RSADecryptHashed decrypts given data with RSA.
func RSADecryptHashed(data []byte, key *rsa.PrivateKey) (r []byte, err error) {
	c := new(big.Int).SetBytes(data)
	m := new(big.Int).Exp(c, key.D, key.N)

	var dataWithHash [rsaWithHashLen]byte
	m.FillBytes(dataWithHash[:])

	hash := dataWithHash[:sha1.Size]
	paddedData := dataWithHash[sha1.Size:]

	// Guessing such data that sha1(data) == hash.
	for i := 0; i <= len(paddedData); i++ {
		data := paddedData[:len(paddedData)-i]
		h := sha1.Sum(data) // #nosec
		if bytes.Equal(h[:], hash) {
			// Found.
			return data, nil
		}
	}

	// This can be caused by invalid keys or implementation bug.
	return nil, xerrors.New("hash mismatch")
}
