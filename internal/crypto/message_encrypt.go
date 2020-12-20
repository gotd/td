package crypto

import (
	"crypto/aes"
	"io"

	"github.com/gotd/td/bin"

	"github.com/gotd/ige"
)

func countPadding(l int) int { return 16 + (16 - (l % 16)) }

// EncryptMessage encrypts plaintext using AES-IGE.
func (c Cipher) EncryptMessage(k AuthKeyWithID, plaintext []byte) (*EncryptedMessage, error) {
	plaintextPadded := make([]byte, len(plaintext)+countPadding(len(plaintext)))
	copy(plaintextPadded, plaintext)
	if _, err := io.ReadFull(c.rand, plaintextPadded[len(plaintext):]); err != nil {
		return nil, err
	}

	messageKey := MessageKey(k.AuthKey, plaintextPadded, c.encryptSide)
	key, iv := Keys(k.AuthKey, messageKey, c.encryptSide)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encrypter := ige.NewIGEEncrypter(cipher, iv[:])
	msg := &EncryptedMessage{
		AuthKeyID:     k.AuthKeyID,
		MsgKey:        messageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	encrypter.CryptBlocks(msg.EncryptedData, plaintextPadded)
	return msg, nil
}

// Encrypt encrypts EncryptedMessageData using AES-IGE to given buffer.
func (c Cipher) Encrypt(key AuthKeyWithID, data EncryptedMessageData, b *bin.Buffer) error {
	b.Reset()
	if err := data.Encode(b); err != nil {
		return err
	}

	msg, err := c.EncryptMessage(key, b.Raw())
	if err != nil {
		return err
	}

	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}

	return nil
}
