package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"math/big"

	"github.com/go-faster/errors"
	"github.com/go-faster/xor"
	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

func reverseBytes(s []byte) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

const (
	rsaPadDataLimit       = 144
	dataWithPaddingLength = 192
	dataWithHashLength    = dataWithPaddingLength + sha256.Size
	tempKeySize           = 32
)

// RSAPad encrypts given data with RSA, prefixing with a hash.
//
// See https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication.
func RSAPad(data []byte, key *rsa.PublicKey, randomSource io.Reader) ([]byte, error) {
	// 1) data_with_padding := data + random_padding_bytes; — where random_padding_bytes are
	// chosen so that the resulting length of data_with_padding is precisely 192 bytes, and
	// data is the TL-serialized data to be encrypted as before.
	//
	// One has to check that data is not longer than 144 bytes.
	if len(data) > rsaPadDataLimit {
		return nil, errors.Errorf("data length is bigger that 144 (%d)", len(data))
	}

	dataWithPadding := make([]byte, dataWithPaddingLength)
	copy(dataWithPadding, data)
	// Filling data_with_padding with random bytes.
	if _, err := io.ReadFull(randomSource, dataWithPadding[len(data):]); err != nil {
		return nil, errors.Wrap(err, "pad data with random")
	}

	// Make a copy.
	dataPadReversed := make([]byte, dataWithPaddingLength)
	copy(dataPadReversed, dataWithPadding)
	// 2) data_pad_reversed := BYTE_REVERSE(data_with_padding);
	reverseBytes(dataPadReversed)

	for {
		// 3) A random 32-byte temp_key is generated.
		tempKey := make([]byte, tempKeySize)
		if _, err := io.ReadFull(randomSource, tempKey); err != nil {
			return nil, errors.Wrap(err, "generate temp_key")
		}

		// 4) data_with_hash := data_pad_reversed + SHA256(temp_key + data_with_padding);
		// — after this assignment, data_with_hash is exactly 224 bytes long.
		dataWithHash := make([]byte, 0, dataWithHashLength)
		dataWithHash = append(dataWithHash, dataPadReversed...)
		{
			h := sha256.New()
			_, _ = h.Write(tempKey)
			_, _ = h.Write(dataWithPadding)
			dataWithHash = h.Sum(dataWithHash)
			dataWithHash = dataWithHash[:dataWithHashLength]
		}

		// 5) aes_encrypted := AES256_IGE(data_with_hash, temp_key, 0); — AES256-IGE encryption with zero IV.
		aesEncrypted := make([]byte, len(dataWithHash))
		{
			aesBlock, err := aes.NewCipher(tempKey)
			if err != nil {
				return nil, errors.Wrap(err, "create cipher")
			}
			var zeroIV bin.Int256
			ige.EncryptBlocks(aesBlock, zeroIV[:], aesEncrypted, dataWithHash)
		}

		// 6) temp_key_xor := temp_key XOR SHA256(aes_encrypted); — adjusted key, 32 bytes
		tempKeyXor := make([]byte, tempKeySize)
		{
			aesEncryptedHash := sha256.Sum256(aesEncrypted)
			xor.Bytes(tempKeyXor, tempKey, aesEncryptedHash[:])
		}

		// 7) key_aes_encrypted := temp_key_xor + aes_encrypted; — exactly 256 bytes (2048 bits) long.
		keyAESEncrypted := make([]byte, 0, tempKeySize+dataWithHashLength)
		keyAESEncrypted = append(keyAESEncrypted, tempKeyXor...)
		keyAESEncrypted = append(keyAESEncrypted, aesEncrypted...)

		// 8) The value of key_aes_encrypted is compared with the RSA-modulus of server_pubkey
		// as a big-endian 2048-bit (256-byte) unsigned integer. If key_aes_encrypted turns out to be
		// greater than or equal to the RSA modulus, the previous steps starting from the generation
		// of new random temp_key are repeated.
		keyAESEncryptedBig := big.NewInt(0).SetBytes(keyAESEncrypted)
		if keyAESEncryptedBig.Cmp(key.N) >= 0 {
			continue
		}
		// Otherwise the final step is performed:

		// 9) encrypted_data := RSA(key_aes_encrypted, server_pubkey);
		// — 256-byte big-endian integer is elevated to the requisite power from the RSA public key
		// modulo the RSA modulus, and the result is stored as a big-endian integer consisting of
		// exactly 256 bytes (with leading zero bytes if required).
		//
		// Encrypting "key_aes_encrypted" with RSA.
		res := rsaEncrypt(keyAESEncrypted, key)
		return res, nil
	}
}

// DecodeRSAPad implements server-side decoder of RSAPad.
func DecodeRSAPad(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	var encryptedData [256]byte
	if !rsaDecrypt(data, key, encryptedData[:]) {
		return nil, errors.New("invalid encrypted_data")
	}

	tempKeyXor := encryptedData[:tempKeySize]
	aesEncrypted := encryptedData[tempKeySize:]

	tempKey := make([]byte, tempKeySize)
	{
		aesEncryptedHash := sha256.Sum256(aesEncrypted)
		xor.Bytes(tempKey, tempKeyXor, aesEncryptedHash[:])
	}

	dataWithHash := make([]byte, len(aesEncrypted))
	{
		aesBlock, err := aes.NewCipher(tempKey)
		if err != nil {
			return nil, errors.Wrap(err, "create cipher")
		}
		var zeroIV bin.Int256
		ige.DecryptBlocks(aesBlock, zeroIV[:], dataWithHash, aesEncrypted)
	}

	dataWithPadding := dataWithHash[:dataWithPaddingLength]
	reverseBytes(dataWithPadding)

	hash := dataWithHash[dataWithPaddingLength:]
	{
		h := sha256.New()
		_, _ = h.Write(tempKey)
		_, _ = h.Write(dataWithPadding)

		if !bytes.Equal(hash, h.Sum(nil)) {
			return nil, errors.New("hash mismatch")
		}
	}

	return dataWithPadding, nil
}
