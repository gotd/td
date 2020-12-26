package crypto

import (
	"crypto/aes"
	"crypto/sha1"
	"errors"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
)

// DecryptExchangeAnswer decrypts messages created during key exchange.
func DecryptExchangeAnswer(data, key, iv []byte) (dst []byte, err error) {
	// Decrypting inner data.
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, xerrors.Errorf("failed to init aes cipher: %w", err)
	}

	d := ige.NewIGEDecrypter(cipher, iv)
	dataWithHash := make([]byte, len(data))
	// Checking length. Invalid length will lead to panic in CryptBlocks.
	if len(dataWithHash)%cipher.BlockSize() != 0 {
		return nil, xerrors.Errorf("invalid len of data_with_hash (%d %% 16 != 0)", len(dataWithHash))
	}
	d.CryptBlocks(dataWithHash, data)

	dst = GuessDataWithHash(dataWithHash)
	if data == nil {
		// Most common cause of this error is invalid crypto implementation,
		// i.e. invalid keys are used to decrypt payload which lead to
		// decrypt failure, so data does not match sha1 with any padding.
		return nil, errors.New("failed to guess data from data_with_hash")
	}

	return
}

// EncryptExchangeAnswer encrypts messages created during key exchange.
func EncryptExchangeAnswer(rand io.Reader, answer, key, iv []byte) (dst []byte, err error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, xerrors.Errorf("failed to init aes cipher: %w", err)
	}

	answerWithHash, err := DataWithHash(answer, rand)
	if err != nil {
		return nil, xerrors.Errorf("failed to get answer with hash: %w", err)
	}

	dst = make([]byte, len(answerWithHash))
	i := ige.NewIGEEncrypter(cipher, iv)
	i.CryptBlocks(dst, answerWithHash)
	return
}

func NonceHash1(newNonce bin.Int256, key AuthKey) (r bin.Int128) {
	var buf []byte
	buf = append(buf, newNonce[:]...)
	buf = append(buf, 1)
	buf = append(buf, sha(key[:])[0:8]...)
	buf = sha(buf)[4:20]
	copy(r[:], buf)
	return
}

func sha(v []byte) []byte {
	h := sha1.Sum(v) // #nosec
	return h[:]
}
