// Code generated by gotdgen, DO NOT EDIT.

package td

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

// SMS represents TL type `sms#ed8bebfe`.
//
// See https://localhost:80/doc/constructor/sms for reference.
type SMS struct {
	// Text field of SMS.
	Text string
}

// SMSTypeID is TL type id of SMS.
const SMSTypeID = 0xed8bebfe

func (s *SMS) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Text == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SMS) String() string {
	if s == nil {
		return "SMS(nil)"
	}
	type Alias SMS
	return fmt.Sprintf("SMS%+v", Alias(*s))
}

// FillFrom fills SMS from given interface.
func (s *SMS) FillFrom(from interface {
	GetText() (value string)
}) {
	s.Text = from.GetText()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SMS) TypeID() uint32 {
	return SMSTypeID
}

// TypeName returns name of type in TL schema.
func (*SMS) TypeName() string {
	return "sms"
}

// TypeInfo returns info about TL type.
func (s *SMS) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "sms",
		ID:   SMSTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Text",
			SchemaName: "text",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SMS) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "sms#ed8bebfe",
		}
	}
	b.PutID(SMSTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SMS) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "sms#ed8bebfe",
		}
	}
	b.PutString(s.Text)
	return nil
}

// GetText returns value of Text field.
func (s *SMS) GetText() (value string) {
	return s.Text
}

// Decode implements bin.Decoder.
func (s *SMS) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "sms#ed8bebfe",
		}
	}
	if err := b.ConsumeID(SMSTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "sms#ed8bebfe",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SMS) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "sms#ed8bebfe",
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "sms#ed8bebfe",
				FieldName:  "text",
				Underlying: err,
			}
		}
		s.Text = value
	}
	return nil
}

// Ensuring interfaces in compile-time for SMS.
var (
	_ bin.Encoder     = &SMS{}
	_ bin.Decoder     = &SMS{}
	_ bin.BareEncoder = &SMS{}
	_ bin.BareDecoder = &SMS{}
)
