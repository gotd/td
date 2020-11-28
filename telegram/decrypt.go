package telegram

import (
	"crypto/aes"

	"go.uber.org/zap"

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
	n := int(msg.MessageDataLen)
	if (n + padding(n)) != len(msg.MessageDataWithPadding) {
		// Probably we don't care?
		c.log.With(
			zap.Int32("message_data_len", msg.MessageDataLen),
			zap.Int("len", len(msg.MessageDataWithPadding)),
			zap.Int("expected_padding", padding(n)),
			zap.Int("expected_length", padding(n)+n),
		).Debug("Invalid padding")
	}

	return msg, nil
}
