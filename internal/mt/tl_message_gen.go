// Code generated by gotdgen, DO NOT EDIT.

package mt

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdp"
	"github.com/gotd/td/tgerr"
)

// No-op definition for keeping imports.
var (
	_ = bin.Buffer{}
	_ = context.Background()
	_ = fmt.Stringer(nil)
	_ = strings.Builder{}
	_ = errors.Is
	_ = multierr.AppendInto
	_ = sort.Ints
	_ = tdp.Format
	_ = tgerr.Error{}
)

// Message represents TL type `message#5bb8e511`.
type Message struct {
	// MsgID field of Message.
	MsgID int64
	// Seqno field of Message.
	Seqno int
	// Bytes field of Message.
	Bytes int
	// Body field of Message.
	Body GzipPacked
}

// MessageTypeID is TL type id of Message.
const MessageTypeID = 0x5bb8e511

func (m *Message) Zero() bool {
	if m == nil {
		return true
	}
	if !(m.MsgID == 0) {
		return false
	}
	if !(m.Seqno == 0) {
		return false
	}
	if !(m.Bytes == 0) {
		return false
	}
	if !(m.Body.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (m *Message) String() string {
	if m == nil {
		return "Message(nil)"
	}
	type Alias Message
	return fmt.Sprintf("Message%+v", Alias(*m))
}

// FillFrom fills Message from given interface.
func (m *Message) FillFrom(from interface {
	GetMsgID() (value int64)
	GetSeqno() (value int)
	GetBytes() (value int)
	GetBody() (value GzipPacked)
}) {
	m.MsgID = from.GetMsgID()
	m.Seqno = from.GetSeqno()
	m.Bytes = from.GetBytes()
	m.Body = from.GetBody()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*Message) TypeID() uint32 {
	return MessageTypeID
}

// TypeName returns name of type in TL schema.
func (*Message) TypeName() string {
	return "message"
}

// TypeInfo returns info about TL type.
func (m *Message) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "message",
		ID:   MessageTypeID,
	}
	if m == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "MsgID",
			SchemaName: "msg_id",
		},
		{
			Name:       "Seqno",
			SchemaName: "seqno",
		},
		{
			Name:       "Bytes",
			SchemaName: "bytes",
		},
		{
			Name:       "Body",
			SchemaName: "body",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (m *Message) Encode(b *bin.Buffer) error {
	if m == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "message#5bb8e511",
		}
	}
	b.PutID(MessageTypeID)
	return m.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (m *Message) EncodeBare(b *bin.Buffer) error {
	if m == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "message#5bb8e511",
		}
	}
	b.PutLong(m.MsgID)
	b.PutInt(m.Seqno)
	b.PutInt(m.Bytes)
	if err := m.Body.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "message#5bb8e511",
			FieldName:  "body",
			Underlying: err,
		}
	}
	return nil
}

// GetMsgID returns value of MsgID field.
func (m *Message) GetMsgID() (value int64) {
	return m.MsgID
}

// GetSeqno returns value of Seqno field.
func (m *Message) GetSeqno() (value int) {
	return m.Seqno
}

// GetBytes returns value of Bytes field.
func (m *Message) GetBytes() (value int) {
	return m.Bytes
}

// GetBody returns value of Body field.
func (m *Message) GetBody() (value GzipPacked) {
	return m.Body
}

// Decode implements bin.Decoder.
func (m *Message) Decode(b *bin.Buffer) error {
	if m == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "message#5bb8e511",
		}
	}
	if err := b.ConsumeID(MessageTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "message#5bb8e511",
			Underlying: err,
		}
	}
	return m.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (m *Message) DecodeBare(b *bin.Buffer) error {
	if m == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "message#5bb8e511",
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "message#5bb8e511",
				FieldName:  "msg_id",
				Underlying: err,
			}
		}
		m.MsgID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "message#5bb8e511",
				FieldName:  "seqno",
				Underlying: err,
			}
		}
		m.Seqno = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "message#5bb8e511",
				FieldName:  "bytes",
				Underlying: err,
			}
		}
		m.Bytes = value
	}
	{
		if err := m.Body.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "message#5bb8e511",
				FieldName:  "body",
				Underlying: err,
			}
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for Message.
var (
	_ bin.Encoder     = &Message{}
	_ bin.Decoder     = &Message{}
	_ bin.BareEncoder = &Message{}
	_ bin.BareDecoder = &Message{}
)
