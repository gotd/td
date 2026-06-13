package crypto

import (
	"crypto/aes"
	"io"

	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

// countPadding returns the amount of random padding (in bytes) to append to a
// plaintext of length l before AES-IGE encryption.
//
// The padding both aligns the plaintext to the 16-byte block size (with at
// least 12 bytes of padding, as required by MTProto 2.0) and adds a random
// number of extra 16-byte blocks. The random component mirrors Telegram
// Desktop's CountPaddingPrimesCount and removes the deterministic
// encrypted-message length that would otherwise fingerprint the client.
//
// randByte provides the entropy for the random component; only its low 4 bits
// are used, yielding 0..15 extra 16-byte blocks (0..240 bytes).
func countPadding(l int, randByte byte) int {
	padding := (16 - (l % 16)) % 16
	if padding < 12 {
		padding += 16
	}
	padding += int(randByte&0x0F) * 16
	return padding
}

// encryptMessage encrypts plaintext using AES-IGE.
func (c Cipher) encryptMessage(k AuthKey, plaintext *bin.Buffer) (EncryptedMessage, error) {
	offset := len(plaintext.Buf)

	var randByte [1]byte
	if _, err := io.ReadFull(c.rand, randByte[:]); err != nil {
		return EncryptedMessage{}, err
	}
	plaintext.Buf = append(plaintext.Buf, make([]byte, countPadding(offset, randByte[0]))...)
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
