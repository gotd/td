package telegram

import (
	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/crypto"
	"github.com/ernado/td/internal/proto"
)

func (c Client) newEncryptedMessage(payload bin.Encoder, b *bin.Buffer) error {
	if err := payload.Encode(b); err != nil {
		return err
	}
	d := proto.EncryptedMessageData{
		SessionID:              c.session,
		Salt:                   c.salt,
		MessageID:              crypto.NewMessageID(c.clock(), crypto.MessageFromClient),
		SeqNo:                  int32(c.seq),
		MessageDataLen:         int32(b.Len()),
		MessageDataWithPadding: b.Copy(),
	}
	b.Reset()
	if err := d.Encode(b); err != nil {
		return err
	}
	msg, err := c.encrypt(b.Copy())
	if err != nil {
		return err
	}
	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}
	c.seq++
	return nil
}
