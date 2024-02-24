package crypto

import (
	"crypto/aes"

	"github.com/go-faster/errors"

	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

// DecryptFromBuffer decodes EncryptedMessage and decrypts it.
func (c Cipher) DecryptFromBuffer(k AuthKey, buf *bin.Buffer) (*EncryptedMessageData, error) {
	msg := &EncryptedMessage{}
	// Because we assume that buffer is valid during decrypting, we able to
	// use DecodeWithoutCopy and do not allocate inner buffer for EncryptedMessage.
	if err := msg.DecodeWithoutCopy(buf); err != nil {
		return nil, err
	}

	return c.Decrypt(k, msg)
}

// Decrypt decrypts data from encrypted message using AES-IGE.
func (c Cipher) Decrypt(k AuthKey, encrypted *EncryptedMessage) (*EncryptedMessageData, error) {
	plaintext, err := c.decryptMessage(k, encrypted)
	if err != nil {
		return nil, err
	}

	side := c.encryptSide.DecryptSide()
	// Checking SHA256 hash value of msg_key
	msgKey := MessageKey(k.Value, plaintext, side)
	if msgKey != encrypted.MsgKey {
		return nil, errors.New("msg_key is invalid")
	}

	msg := &EncryptedMessageData{}
	// Notice: do not re-use plaintext, because we use DecodeWithoutCopy, it references
	// original buffer.
	if err := msg.DecodeWithoutCopy(&bin.Buffer{Buf: plaintext}); err != nil {
		return nil, err
	}

	{
		// Checking that padding of decrypted message is not too big.
		const maxPadding = 1024
		n := int(msg.MessageDataLen)
		paddingLen := len(msg.MessageDataWithPadding) - n

		switch {
		case n < 0:
			return nil, errors.Errorf("message length is invalid: %d less than zero", n)
		case n%4 != 0:
			return nil, errors.Errorf("message length is invalid: %d is not divisible by 4", n)
		case paddingLen > maxPadding:
			return nil, errors.Errorf("padding %d of message is too big", paddingLen)
		}
	}

	return msg, nil
}

// decryptMessage decrypts data from encrypted message using AES-IGE.
func (c Cipher) decryptMessage(k AuthKey, encrypted *EncryptedMessage) ([]byte, error) {
	if k.ID != encrypted.AuthKeyID {
		return nil, errors.New("unknown auth key id")
	}
	if len(encrypted.EncryptedData)%16 != 0 {
		return nil, errors.New("invalid encrypted data padding")
	}

	key, iv := Keys(k.Value, encrypted.MsgKey, c.encryptSide.DecryptSide())
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(encrypted.EncryptedData))
	ige.DecryptBlocks(cipher, iv[:], plaintext, encrypted.EncryptedData)

	return plaintext, nil
}
