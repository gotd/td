package proto

import (
	"fmt"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"
)

// UnencryptedMessage is plaintext message.
type UnencryptedMessage struct {
	MessageID   crypto.MessageID
	MessageData []byte
}

func (u *UnencryptedMessage) Decode(b *bin.Buffer) error {
	{
		id, err := b.Long()
		if err != nil {
			return err
		}
		if id != 0 {
			return fmt.Errorf("unexpected auth_key_id %d of plaintext message", id)
		}
	}
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		u.MessageID = crypto.MessageID(v)
	}

	// Reading data.
	dataLen, err := b.Int32()
	if err != nil {
		return err
	}
	u.MessageData = append(u.MessageData[:0], make([]byte, dataLen)...)
	if err := b.ConsumeN(u.MessageData, int(dataLen)); err != nil {
		return fmt.Errorf("failed to consume payload: %w", err)
	}

	return nil
}

func (u UnencryptedMessage) Encode(b *bin.Buffer) error {
	b.PutLong(0)
	b.PutLong(int64(u.MessageID))
	b.PutInt32(int32(len(u.MessageData)))
	b.Put(u.MessageData)
	return nil
}
