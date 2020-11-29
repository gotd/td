package telegram

import (
	"crypto/aes"

	"github.com/ernado/ige"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) decrypt(encrypted *proto.EncryptedMessage) ([]byte, error) {
	if c.authKeyID != encrypted.AuthKeyID {
		return nil, xerrors.New("unknown auth key id")
	}
	if len(encrypted.EncryptedData)%16 != 0 {
		return nil, xerrors.New("invalid encrypted data padding")
	}

	key, iv := crypto.Keys(c.authKey, encrypted.MsgKey, crypto.Server)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(encrypted.EncryptedData))
	d := ige.NewIGEDecrypter(cipher, iv[:])
	d.CryptBlocks(plaintext, encrypted.EncryptedData)

	return plaintext, nil
}

func (c *Client) decryptData(encrypted *proto.EncryptedMessage) (*proto.EncryptedMessageData, error) {
	plaintext, err := c.decrypt(encrypted)
	if err != nil {
		return nil, err
	}
	b := &bin.Buffer{Buf: plaintext}
	msg := &proto.EncryptedMessageData{}
	if err := msg.Decode(b); err != nil {
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
