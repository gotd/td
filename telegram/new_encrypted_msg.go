package telegram

import (
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) newEncryptedMessage(id crypto.MessageID, seq int32, payload bin.Encoder, b *bin.Buffer) error {
	if err := payload.Encode(b); err != nil {
		return err
	}
	d := proto.EncryptedMessageData{
		SessionID:              atomic.LoadInt64(&c.session),
		Salt:                   atomic.LoadInt64(&c.salt),
		MessageID:              id,
		SeqNo:                  seq,
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

	c.log.With(zap.Int64("request_id", int64(d.MessageID))).Debug("Request")

	return nil
}
