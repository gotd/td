package proto

import (
	"errors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/crypto"

	"golang.org/x/xerrors"
)

const MessageContainerTypeID = 0x73f1f8dc

type MessageContainer struct {
	Messages []Message
}

func (m *MessageContainer) Decode(b *bin.Buffer) error {
	if err := b.ConsumeID(MessageContainerTypeID); err != nil {
		return xerrors.Errorf("failed to consume id of message container: %w", err)
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

type Message struct {
	ID    crypto.MessageID
	SeqNo int
	Bytes int
	Body  []byte
}

func (m *Message) Decode(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		m.ID = crypto.MessageID(v)
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
