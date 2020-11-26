package telegram

import (
	"crypto/aes"
	"io"

	"github.com/ernado/ige"

	"github.com/ernado/td/internal/crypto"
	"github.com/ernado/td/internal/proto"
)

func padding(l int) int { return 16 + (16 - (l % 16)) }

// encrypt encrypts plaintext using AES-IGE.
func (c Client) encrypt(plaintext []byte) (*proto.EncryptedMessage, error) {
	plaintextPadded := make([]byte, len(plaintext)+padding(len(plaintext)))
	copy(plaintextPadded, plaintext)
	if _, err := io.ReadFull(c.rand, plaintextPadded[len(plaintext):]); err != nil {
		return nil, err
	}
	keys := crypto.MessageKeys(c.authKey, plaintextPadded, crypto.Client)
	cipher, err := aes.NewCipher(keys.Key[:])
	if err != nil {
		return nil, err
	}
	encryptor := ige.NewIGEEncrypter(cipher, keys.IV[:])
	msg := &proto.EncryptedMessage{
		AuthKeyID:     c.authKeyID,
		MsgKey:        keys.MessageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	encryptor.CryptBlocks(msg.EncryptedData, plaintextPadded)
	return msg, nil
}
