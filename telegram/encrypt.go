package telegram

import (
	"crypto/aes"
	"io"

	"github.com/gotd/ige"
	"github.com/gotd/td/internal/crypto"
)

func padding(l int) int { return 16 + (16 - (l % 16)) }

// encrypt encrypts plaintext using AES-IGE.
func (c *Client) encrypt(plaintext []byte) (*crypto.EncryptedMessage, error) {
	plaintextPadded := make([]byte, len(plaintext)+padding(len(plaintext)))
	copy(plaintextPadded, plaintext)
	if _, err := io.ReadFull(c.rand, plaintextPadded[len(plaintext):]); err != nil {
		return nil, err
	}

	messageKey := crypto.MessageKey(c.authKey, plaintextPadded, crypto.Client)
	key, iv := crypto.Keys(c.authKey, messageKey, crypto.Client)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encryptor := ige.NewIGEEncrypter(cipher, iv[:])
	msg := &crypto.EncryptedMessage{
		AuthKeyID:     c.authKeyID,
		MsgKey:        messageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	encryptor.CryptBlocks(msg.EncryptedData, plaintextPadded)
	return msg, nil
}
