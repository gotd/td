package crypto

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505
	"io"

	"github.com/go-faster/errors"
)

// RSAEncryptHashed encrypts given data with RSA, prefixing with a hash.
func RSAEncryptHashed(data []byte, key *rsa.PublicKey, randomSource io.Reader) ([]byte, error) {
	// Preparing `data_with_hash`.
	// data_with_hash := SHA1(data) + data + (any random bytes);
	// such that the length equals 255 bytes;
	var dataWithHash [rsaWithHashLen]byte
	if len(data) > rsaDataLen {
		return nil, errors.Errorf("data length %d is too big", len(data))
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
	res := rsaEncrypt(dataWithHash[:], key)

	return res, nil
}

// RSADecryptHashed decrypts given data with RSA.
func RSADecryptHashed(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	var dataWithHash [rsaWithHashLen]byte
	if !rsaDecrypt(data, key, dataWithHash[:]) {
		return nil, errors.New("invalid data_with_hash")
	}

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
	return nil, errors.New("hash mismatch")
}
