package crypto

import (
	"crypto/aes"
	"io"

	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

func countPadding(l int) int { return 16 + (16 - (l % 16)) }

// encryptMessage encrypts plaintext using AES-IGE.
func (c Cipher) encryptMessage(k AuthKey, plaintext *bin.Buffer) (EncryptedMessage, error) {
	offset := len(plaintext.Buf)
	plaintext.Buf = append(plaintext.Buf, make([]byte, countPadding(offset))...)
	if _, err := io.ReadFull(c.rand, plaintext.Buf[offset:]); err != nil {
		return EncryptedMessage{}, err
	}

	messageKey := MessageKey(k.Value, plaintext.Buf, c.encryptSide)
	key, iv := Keys(k.Value, messageKey, c.encryptSide)
	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return EncryptedMessage{}, err
	}
	msg := EncryptedMessage{
		AuthKeyID:     k.ID,
		MsgKey:        messageKey,
		EncryptedData: make([]byte, len(plaintext.Buf)),
	}
	ige.EncryptBlocks(aesBlock, iv[:], msg.EncryptedData, plaintext.Buf)
	return msg, nil
}

// Encrypt encrypts EncryptedMessageData using AES-IGE to given buffer.
func (c Cipher) Encrypt(key AuthKey, data EncryptedMessageData, b *bin.Buffer) error {
	b.Reset()
	if err := data.EncodeWithoutCopy(b); err != nil {
		return err
	}

	msg, err := c.encryptMessage(key, b)
	if err != nil {
		return err
	}

	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}

	return nil
}
