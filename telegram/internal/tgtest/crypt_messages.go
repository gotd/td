package tgtest

import (
	"crypto/aes"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

func padding(l int) int { return 16 + (16 - (l % 16)) }

// encrypt encrypts plaintext using AES-IGE.
func (s *Server) encrypt(authKey crypto.AuthKey, plaintext []byte) (*crypto.EncryptedMessage, error) {
	plaintextPadded := make([]byte, len(plaintext)+padding(len(plaintext)))
	copy(plaintextPadded, plaintext)
	if _, err := io.ReadFull(s.rand, plaintextPadded[len(plaintext):]); err != nil {
		return nil, err
	}

	messageKey := crypto.MessageKey(authKey, plaintextPadded, crypto.Server)
	key, iv := crypto.Keys(authKey, messageKey, crypto.Server)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	encryptor := ige.NewIGEEncrypter(cipher, iv[:])
	msg := &crypto.EncryptedMessage{
		AuthKeyID:     authKey.ID(),
		MsgKey:        messageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	encryptor.CryptBlocks(msg.EncryptedData, plaintextPadded)
	return msg, nil
}

func (s *Server) encryptData(authKey crypto.AuthKey, b bin.Encoder) (*crypto.EncryptedMessage, error) {
	var buf bin.Buffer

	if err := b.Encode(&buf); err != nil {
		return nil, err
	}

	data := &crypto.EncryptedMessageData{
		MessageDataLen:         int32(buf.Len()),
		MessageDataWithPadding: buf.Copy(),
	}

	buf.Reset()
	if err := data.Encode(&buf); err != nil {
		return nil, err
	}

	return s.encrypt(authKey, buf.Raw())
}

func (s *Server) decrypt(authKey crypto.AuthKey, encrypted *crypto.EncryptedMessage) ([]byte, error) {
	if authKey.ID() != encrypted.AuthKeyID {
		return nil, xerrors.New("unknown auth key id")
	}
	if len(encrypted.EncryptedData)%16 != 0 {
		return nil, xerrors.New("invalid encrypted data padding")
	}

	key, iv := crypto.Keys(authKey, encrypted.MsgKey, crypto.Client)
	cipher, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(encrypted.EncryptedData))
	d := ige.NewIGEDecrypter(cipher, iv[:])
	d.CryptBlocks(plaintext, encrypted.EncryptedData)

	return plaintext, nil
}

func (s *Server) decryptData(authKey crypto.AuthKey, encrypted *crypto.EncryptedMessage) (*crypto.EncryptedMessageData, error) {
	plaintext, err := s.decrypt(authKey, encrypted)
	if err != nil {
		return nil, err
	}
	b := &bin.Buffer{Buf: plaintext}
	msg := &crypto.EncryptedMessageData{}
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
