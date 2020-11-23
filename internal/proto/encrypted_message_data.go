package proto

import (
	"fmt"

	"github.com/ernado/td/internal/crypto"

	"github.com/ernado/td/bin"
)

// EncryptedMessageData is stored in EncryptedMessage.EncryptedData.
type EncryptedMessageData struct {
	Salt        int64
	SessionID   int64
	MessageID   crypto.MessageID
	SeqNo       int32
	MessageData []byte
}

// encryptedMessageDataHeader is 3 int64 + 2 int32.
const encryptedMessageDataHeader = 3*2*bin.Word + 2*bin.Word

// Encode implements bin.Encoder.
func (e EncryptedMessageData) Encode(b *bin.Buffer) error {
	b.PutLong(e.Salt)
	b.PutLong(e.SessionID)
	b.PutLong(int64(e.MessageID))
	b.PutInt32(e.SeqNo)
	b.PutInt32(int32(len(e.MessageData)))
	b.Put(e.MessageData)
	b.PutPadding(paddingRequired(len(e.MessageData) + encryptedMessageDataHeader))
	return nil
}

// Decode implements bin.Decoder.
func (e *EncryptedMessageData) Decode(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		e.Salt = v
	}
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		e.SessionID = v
	}
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		e.MessageID = crypto.MessageID(v)
	}
	{
		v, err := b.Int32()
		if err != nil {
			return err
		}
		e.SeqNo = v
	}

	// Reading data.
	dataLen, err := b.Int32()
	if err != nil {
		return err
	}
	e.MessageData = append(e.MessageData[:0], make([]byte, dataLen)...)
	if err := b.ConsumeN(e.MessageData, int(dataLen)); err != nil {
		return fmt.Errorf("failed to consume payload: %w", err)
	}
	if err := b.ConsumePadding(paddingRequired(int(dataLen) + encryptedMessageDataHeader)); err != nil {
		return fmt.Errorf("failed to consume padding: %w", err)
	}
	return nil
}
