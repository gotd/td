package crypto

import (
	"crypto/aes"

	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
)

// DecryptMessage decrypts data from encrypted message using AES-IGE.
func (c Cipher) DecryptMessage(authKey AuthKey, encrypted *EncryptedMessage) ([]byte, error) {
	if authKey.ID() != encrypted.AuthKeyID {
		return nil, xerrors.New("unknown auth key id")
	}
	if len(encrypted.EncryptedData)%16 != 0 {
		return nil, xerrors.New("invalid encrypted data padding")
	}

	key, iv := Keys(authKey, encrypted.MsgKey, c.encryptSide.DecryptSide())
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(encrypted.EncryptedData))
	d := ige.NewIGEDecrypter(cipher, iv[:])
	d.CryptBlocks(plaintext, encrypted.EncryptedData)

	return plaintext, nil
}

// DecryptDataFrom decrypts data from buffer with EncryptedMessage using AES-IGE.
func (c Cipher) DecryptDataFrom(authKey AuthKey, sessionID int64, b *bin.Buffer) (*EncryptedMessageData, error) {
	encrypted := &EncryptedMessage{}
	if err := encrypted.Decode(b); err != nil {
		return nil, xerrors.Errorf("failed to decode encrypted message: %w", err)
	}

	plaintext, err := c.DecryptMessage(authKey, encrypted)
	if err != nil {
		return nil, err
	}

	side := c.encryptSide.DecryptSide()
	// Checking SHA256 hash value of msg_key
	key := MessageKey(authKey, plaintext, side)
	if key != encrypted.MsgKey {
		return nil, xerrors.Errorf("msg_key is invalid")
	}

	b.ResetTo(plaintext)
	msg := &EncryptedMessageData{}

	if err := msg.Decode(b); err != nil {
		return nil, err
	}

	// Checking session_id
	//
	// The client is to check that the session_id field in the decrypted message indeed
	// equals to that of an active session created by the client.
	if side != Client && msg.SessionID != sessionID { // Skip check on client.
		return nil, xerrors.Errorf("session id is invalid")
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
