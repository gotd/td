package telegram

import (
	"crypto/aes"
	"fmt"
	"io"

	"github.com/ernado/ige"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/crypto"
	"github.com/ernado/td/internal/proto"
)

func paddedLen16(l int) int {
	n := 16 * (l / 16)
	if n < l {
		n += 16
	}
	return n
}

func (c Client) newEncryptedMessage(payload bin.Encoder, b *bin.Buffer) error {
	if err := payload.Encode(b); err != nil {
		return err
	}
	d := proto.EncryptedMessageData{
		SessionID:              c.session,
		Salt:                   c.salt,
		MessageID:              crypto.NewMessageID(c.clock(), crypto.MessageFromClient),
		SeqNo:                  0,
		MessageDataLen:         int32(len(b.Buf)),
		MessageDataWithPadding: append([]byte{}, b.Buf...),
	}
	b.Reset()
	if err := d.Encode(b); err != nil {
		return err
	}
	plaintextPadded := make([]byte, paddedLen16(len(b.Buf)))
	copy(plaintextPadded, b.Buf)
	if _, err := io.ReadFull(c.rand, b.Buf[len(b.Buf):]); err != nil {
		return err
	}
	keys := crypto.MessageKeys(c.authKey, plaintextPadded, crypto.ModeClient)
	cipher, err := aes.NewCipher(keys.Key[:])
	if err != nil {
		return err
	}
	encryptor := ige.NewIGEEncrypter(cipher, keys.IV[:])
	msg := proto.EncryptedMessage{
		AuthKeyID:     c.authKeyID,
		MsgKey:        keys.MessageKey,
		EncryptedData: make([]byte, len(plaintextPadded)),
	}
	fmt.Println("auth_key_id", msg.AuthKeyID)
	encryptor.CryptBlocks(msg.EncryptedData, plaintextPadded)
	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}
	return nil
}
