package proto

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// MessageContainerTypeID is TL type id of MessageContainer.
const MessageContainerTypeID = 0x73f1f8dc

// MessageContainer contains slice of messages.
type MessageContainer struct {
	Messages []Message
}

// Encode implements bin.Decoder.
func (m *MessageContainer) Encode(b *bin.Buffer) error {
	b.PutID(MessageContainerTypeID)
	b.PutInt(len(m.Messages))
	for _, msg := range m.Messages {
		if err := msg.Encode(b); err != nil {
			return err
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (m *MessageContainer) Decode(b *bin.Buffer) error {
	if err := b.ConsumeID(MessageContainerTypeID); err != nil {
		return errors.Wrap(err, "consume id of message container")
	}
	n, err := b.Int()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		var msg Message
		if err := msg.Decode(b); err != nil {
			return err
		}
		m.Messages = append(m.Messages, msg)
	}
	return nil
}

// Message is element of MessageContainer.
type Message struct {
	ID    int64
	SeqNo int
	Bytes int
	Body  []byte
}

// Encode implements bin.Encoder.
func (m *Message) Encode(b *bin.Buffer) error {
	if m.Bytes < 0 || m.Bytes > 1024*1024 {
		return errors.Errorf("message length %d is invalid", m.Bytes)
	}
	b.PutLong(m.ID)
	b.PutInt(m.SeqNo)
	b.PutInt(m.Bytes)
	b.Put(m.Body)
	return nil
}

// Decode implements bin.Decoder.
func (m *Message) Decode(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		m.ID = v
	}
	{
		v, err := b.Int()
		if err != nil {
			return err
		}
		m.SeqNo = v
	}
	{
		v, err := b.Int()
		if err != nil {
			return err
		}
		m.Bytes = v
	}
	if m.Bytes < 0 || m.Bytes > 1024*1024 {
		return errors.New("message length is too big")
	}
	m.Body = make([]byte, m.Bytes)
	return b.ConsumeN(m.Body, m.Bytes)
}
