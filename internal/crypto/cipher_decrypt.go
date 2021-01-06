package crypto

import (
	"crypto/aes"

	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
)

// DecryptFromBuffer decodes EncryptedMessage and decrypts it.
func (c Cipher) DecryptFromBuffer(k AuthKey, buf *bin.Buffer) (*EncryptedMessageData, error) {
	msg := &EncryptedMessage{}
	if err := msg.Decode(buf); err != nil {
		return nil, err
	}

	return c.Decrypt(k, msg)
}

// Decrypt decrypts data from encrypted message using AES-IGE.
func (c Cipher) Decrypt(k AuthKey, encrypted *EncryptedMessage) (*EncryptedMessageData, error) {
	plaintext, err := c.DecryptMessage(k, encrypted)
	if err != nil {
		return nil, err
	}

	side := c.encryptSide.DecryptSide()
	// Checking SHA256 hash value of msg_key
	msgKey := MessageKey(k.Value, plaintext, side)
	if msgKey != encrypted.MsgKey {
		return nil, xerrors.Errorf("msg_key is invalid")
	}

	msg := &EncryptedMessageData{}
	if err := msg.Decode(&bin.Buffer{Buf: plaintext}); err != nil {
		return nil, err
	}

	{
		// Checking that padding of decrypted message is not too big.
		const maxPadding = 1024
		n := int(msg.MessageDataLen)
		paddingLen := len(msg.MessageDataWithPadding) - n
		if paddingLen > maxPadding {
			return nil, xerrors.Errorf("padding %d of message is too big", paddingLen)
		}
	}

	return msg, nil
}

// DecryptMessage decrypts data from encrypted message using AES-IGE.
func (c Cipher) DecryptMessage(k AuthKey, encrypted *EncryptedMessage) ([]byte, error) {
	if k.ID != encrypted.AuthKeyID {
		return nil, xerrors.New("unknown auth key id")
	}
	if len(encrypted.EncryptedData)%16 != 0 {
		return nil, xerrors.New("invalid encrypted data padding")
	}

	key, iv := Keys(k.Value, encrypted.MsgKey, c.encryptSide.DecryptSide())
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(encrypted.EncryptedData))
	d := ige.NewIGEDecrypter(cipher, iv[:])
	d.CryptBlocks(plaintext, encrypted.EncryptedData)

	return plaintext, nil
}
