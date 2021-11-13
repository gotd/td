package proto

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// UnencryptedMessage is plaintext message.
type UnencryptedMessage struct {
	MessageID   int64
	MessageData []byte
}

// Decode implements bin.Decoder.
func (u *UnencryptedMessage) Decode(b *bin.Buffer) error {
	{
		// Reading auth_key_id that should be always equal to zero.
		id, err := b.Long()
		if err != nil {
			return err
		}
		if id != 0 {
			return errors.Errorf("unexpected auth_key_id %d of plaintext message", id)
		}
	}
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		u.MessageID = v
	}

	// Reading data.
	dataLen, err := b.Int32()
	if err != nil {
		return err
	}
	u.MessageData = append(u.MessageData[:0], make([]byte, dataLen)...)
	if err := b.ConsumeN(u.MessageData, int(dataLen)); err != nil {
		return errors.Wrap(err, "consume payload")
	}

	return nil
}

// Encode implements bin.Encoder.
func (u UnencryptedMessage) Encode(b *bin.Buffer) error {
	b.PutLong(0)
	b.PutLong(u.MessageID)
	b.PutInt32(int32(len(u.MessageData)))
	b.Put(u.MessageData)
	return nil
}
