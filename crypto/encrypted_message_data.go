package crypto

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// EncryptedMessageData is stored in EncryptedMessage.EncryptedData.
type EncryptedMessageData struct {
	Salt                   int64
	SessionID              int64
	MessageID              int64
	SeqNo                  int32
	MessageDataLen         int32
	MessageDataWithPadding []byte

	// Message to encode to MessageDataWithPadding.
	// Needed to prevent unnecessary allocations in EncodeWithoutCopy.
	Message bin.Encoder
}

// Encode implements bin.Encoder.
func (e EncryptedMessageData) Encode(b *bin.Buffer) error {
	b.PutLong(e.Salt)
	b.PutLong(e.SessionID)
	b.PutLong(e.MessageID)
	b.PutInt32(e.SeqNo)
	b.PutInt32(e.MessageDataLen)
	b.Put(e.MessageDataWithPadding)
	return nil
}

// EncodeWithoutCopy is like Encode, but tries to encode Message and uses only one buffer
// to encode. If Message is nil, fallbacks to Encode.
func (e EncryptedMessageData) EncodeWithoutCopy(b *bin.Buffer) error {
	if e.Message == nil {
		return e.Encode(b)
	}

	b.PutLong(e.Salt)
	b.PutLong(e.SessionID)
	b.PutLong(e.MessageID)
	b.PutInt32(e.SeqNo)
	lengthOffset := b.Len()
	b.PutInt32(0)
	originalLength := b.Len()
	if err := b.Encode(e.Message); err != nil {
		return errors.Wrap(err, "encode inner message")
	}
	msgLen := b.Len() - originalLength

	(&bin.Buffer{Buf: b.Buf[lengthOffset:lengthOffset]}).PutInt(msgLen)
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
		e.MessageID = v
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
	if int(e.MessageDataLen) > len(e.MessageDataWithPadding) {
		return errors.Errorf(
			"MessageDataLen field is bigger then MessageDataWithPadding length (%d > %d)",
			int(e.MessageDataLen), len(e.MessageDataWithPadding),
		)
	}

	return nil
}

// DecodeWithoutCopy is like Decode, but MessageDataWithPadding references to given buffer instead of
// copying.
func (e *EncryptedMessageData) DecodeWithoutCopy(b *bin.Buffer) error {
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
		e.MessageID = v
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
	e.MessageDataWithPadding = b.Buf
	if int(e.MessageDataLen) > len(e.MessageDataWithPadding) {
		return errors.Errorf(
			"MessageDataLen field is bigger then MessageDataWithPadding length (%d > %d)",
			int(e.MessageDataLen), len(e.MessageDataWithPadding),
		)
	}
	return nil
}

// Data returns message data without hash.
func (e *EncryptedMessageData) Data() []byte {
	return e.MessageDataWithPadding[:e.MessageDataLen]
}
