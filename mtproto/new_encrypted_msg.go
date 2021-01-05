package mtproto

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

func (c *Conn) newEncryptedMessage(id int64, seq int32, payload bin.Encoder, b *bin.Buffer) error {
	if err := payload.Encode(b); err != nil {
		return err
	}
	{
		typeID, err := b.PeekID()
		if err == nil {
			c.log.With(
				zap.Int64("msg_id", id),
				zap.String("message_type", fmt.Sprintf("0x%x", typeID)),
				zap.String("message_type_str", c.types.Get(typeID)),
			).Debug("Request")
		}
	}
	s := c.session()
	d := crypto.EncryptedMessageData{
		SessionID:              s.ID,
		Salt:                   s.Salt,
		MessageID:              id,
		SeqNo:                  seq,
		MessageDataLen:         int32(b.Len()),
		MessageDataWithPadding: b.Copy(),
	}
	if err := c.cipher.Encrypt(s.Key, d, b); err != nil {
		return err
	}

	return nil
}
