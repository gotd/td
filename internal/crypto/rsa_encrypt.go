package crypto

import (
	"crypto/rsa"
	"fmt"
	"io"
	"math/big"

	// #nosec
	//
	// Allowing sha1 because it is used in MTProto itself.
	"crypto/sha1"
)

// EncryptHashed encrypts given data with RSA, prefixing with a hash.
func EncryptHashed(data []byte, key *rsa.PublicKey, randomSource io.Reader) ([]byte, error) {
	// Preparing `data_with_hash`.
	// data_with_hash := SHA1(data) + data + (any random bytes);
	// such that the length equals 255 bytes;
	var dataWithHash = [255]byte{}
	if len(data) > len(dataWithHash)+sha1.Size {
		return nil, fmt.Errorf("data length %d is too big", len(data))
	}

	// Filling data_with_hash with random bytes.
	if _, err := io.ReadFull(randomSource, dataWithHash[:]); err != nil {
		return nil, err
	}
	h := sha1.New() // #nosec
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	// Replacing first 20 bytes with sha1(data).
	copy(dataWithHash[:sha1.Size], h.Sum(nil))
	// Replacing other bytes with data itself.
	copy(dataWithHash[sha1.Size:], data)

	// Encrypting "dataWithHash" with RSA.
	z := big.NewInt(0).SetBytes(dataWithHash[:])
	e := big.NewInt(int64(key.E))
	c := big.NewInt(0).Exp(z, e, key.N)
	res := make([]byte, 256)
	copy(res, c.Bytes())

	return res, nil
}
