package proto

import (
	"github.com/ernado/td/crypto"

	"github.com/ernado/td/bin"
)

// EncryptedMessageData is stored in EncryptedMessage.EncryptedData.
type EncryptedMessageData struct {
	Salt                   int64
	SessionID              int64
	MessageID              crypto.MessageID
	SeqNo                  int32
	MessageDataLen         int32
	MessageDataWithPadding []byte
}

// Encode implements bin.Encoder.
func (e EncryptedMessageData) Encode(b *bin.Buffer) error {
	b.PutLong(e.Salt)
	b.PutLong(e.SessionID)
	b.PutLong(int64(e.MessageID))
	b.PutInt32(e.SeqNo)
	b.PutInt32(e.MessageDataLen)
	b.Put(e.MessageDataWithPadding)
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
	{
		v, err := b.Int32()
		if err != nil {
			return err
		}
		e.MessageDataLen = v
	}
	e.MessageDataWithPadding = append(e.MessageDataWithPadding[:0], b.Buf...)
	return nil
}
