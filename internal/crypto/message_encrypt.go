package crypto

import (
	"crypto/aes"
	"io"

	"github.com/gotd/td/bin"

	"github.com/gotd/ige"
)

func countPadding(l int) int { return 16 + (16 - (l % 16)) }

// EncryptMessage encrypts plaintext using AES-IGE.
func (c Cipher) EncryptMessage(authKey AuthKey, plaintext []byte) (*EncryptedMessage, error) {
	plaintextPadded := make([]byte, len(plaintext)+countPadding(len(plaintext)))
	copy(plaintextPadded, plaintext)
	if _, err := io.ReadFull(c.rand, plaintextPadded[len(plaintext):]); err != nil {
		return nil, err
	}

	messageKey := MessageKey(authKey, plaintextPadded, c.encryptSide)
	key, iv := Keys(authKey, messageKey, c.encryptSide)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encryptor := ige.NewIGEEncrypter(cipher, iv[:])
	msg := &EncryptedMessage{
		AuthKeyID:     authKey.ID(),
		MsgKey:        messageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	encryptor.CryptBlocks(msg.EncryptedData, plaintextPadded)
	return msg, nil
}

// EncryptDataTo encrypts EncryptedMessageData using AES-IGE to given buffer.
func (c Cipher) EncryptDataTo(authKey AuthKey, data EncryptedMessageData, b *bin.Buffer) error {
	b.Reset()
	if err := data.Encode(b); err != nil {
		return err
	}

	msg, err := c.EncryptMessage(authKey, b.Raw())
	if err != nil {
		return err
	}

	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}

	return nil
}
