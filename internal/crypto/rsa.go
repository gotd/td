package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/rsa"
	// #nosec
	//
	// Allowing sha1 because it is used in MTProto itself.
	"crypto/sha1"
	"crypto/sha256"
	"io"
	"math/big"

	"github.com/gotd/ige"
	"github.com/gotd/xor"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
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
	res := rsaEncrypt(dataWithHash[:], key)

	return res, nil
}

func reverseBytes(s []byte) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// RSAPad encrypts given data with RSA, prefixing with a hash.
//
// See https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication.
func RSAPad(data []byte, key *rsa.PublicKey, randomSource io.Reader) ([]byte, error) {
	const (
		dataWithPaddingLength = 192
		dataWithHashLength    = 224
	)

	// 1) data_with_padding := data + random_padding_bytes; — where random_padding_bytes are
	// chosen so that the resulting length of data_with_padding is precisely 192 bytes, and
	// data is the TL-serialized data to be encrypted as before.
	//
	// One has to check that data is not longer than 144 bytes.
	if len(data) > 144 {
		return nil, xerrors.Errorf("data length is bigger that 144 (%d)", len(data))
	}

	dataWithPadding := make([]byte, dataWithPaddingLength)
	copy(dataWithPadding, data)
	// Filling data_with_padding with random bytes.
	if _, err := io.ReadFull(randomSource, dataWithPadding[len(data):]); err != nil {
		return nil, xerrors.Errorf("pad data with random: %w", err)
	}

	// Make a copy.
	dataPadReversed := append([]byte(nil), dataWithPadding...)
	// 2) data_pad_reversed := BYTE_REVERSE(data_with_padding);
	reverseBytes(dataPadReversed)

	for {
		// 3) A random 32-byte temp_key is generated.
		tempKey := make([]byte, 32)
		if _, err := io.ReadFull(randomSource, tempKey); err != nil {
			return nil, xerrors.Errorf("generate temp_key: %w", err)
		}

		// 4) data_with_hash := data_pad_reversed + SHA256(temp_key + data_with_padding);
		// — after this assignment, data_with_hash is exactly 224 bytes long.
		dataWithHash := make([]byte, 0, dataWithHashLength)
		dataWithHash = append(dataWithHash, dataPadReversed...)
		{
			h := sha256.New()
			_, _ = h.Write(tempKey)
			_, _ = h.Write(dataWithPadding)
			dataWithHash = append(dataWithHash, h.Sum(nil)...)
		}

		// 5) aes_encrypted := AES256_IGE(data_with_hash, temp_key, 0); — AES256-IGE encryption with zero IV.
		aesEncrypted := make([]byte, len(dataWithHash))
		{
			aesBlock, err := aes.NewCipher(tempKey)
			if err != nil {
				return nil, xerrors.Errorf("create cipher: %w", err)
			}
			var zeroIV bin.Int256
			ige.EncryptBlocks(aesBlock, zeroIV[:], aesEncrypted, dataWithHash)
		}

		// 6) temp_key_xor := temp_key XOR SHA256(aes_encrypted); — adjusted key, 32 bytes
		tempKeyXor := make([]byte, 32)
		{
			aesEncryptedHash := sha256.Sum256(aesEncrypted)
			xor.Bytes(tempKeyXor, tempKey, aesEncryptedHash[:])
		}

		// 7) key_aes_encrypted := temp_key_xor + aes_encrypted; — exactly 256 bytes (2048 bits) long.
		keyAESEncrypted := make([]byte, 0, 256)
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
