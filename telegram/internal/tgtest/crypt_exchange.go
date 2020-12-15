package tgtest

import (
	"crypto/aes"
	"errors"

	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

func (s *Server) decryptedExchangeAnswer(data []byte, newNonce bin.Int256, serverNonce bin.Int128) (dst []byte, err error) {
	key, iv := crypto.TempAESKeys(newNonce.BigInt(), serverNonce.BigInt())
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

	dst = crypto.GuessDataWithHash(dataWithHash)
	if data == nil {
		// Most common cause of this error is invalid crypto implementation,
		// i.e. invalid keys are used to decrypt payload which lead to
		// decrypt failure, so data does not match sha1 with any padding.
		return nil, errors.New("failed to guess data from data_with_hash")
	}

	return
}

func (s *Server) encryptedExchangeAnswer(answer []byte, newNonce bin.Int256, serverNonce bin.Int128) (dst []byte, err error) {
	key, iv := crypto.TempAESKeys(newNonce.BigInt(), serverNonce.BigInt())
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, xerrors.Errorf("failed to init aes cipher: %w", err)
	}

	answerWithHash, err := crypto.DataWithHash(answer, s.rand)
	if err != nil {
		return nil, xerrors.Errorf("failed to get answer with hash: %w", err)
	}

	dst = make([]byte, len(answerWithHash))
	i := ige.NewIGEEncrypter(cipher, iv)
	i.CryptBlocks(dst, answerWithHash)
	return
}
