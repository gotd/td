// Code generated by gotdgen, DO NOT EDIT.

package tg

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

// HelpSupport represents TL type `help.support#17c6b5f6`.
// Info on support user.
//
// See https://core.telegram.org/constructor/help.support for reference.
type HelpSupport struct {
	// Phone number
	PhoneNumber string
	// User
	User UserClass
}

// HelpSupportTypeID is TL type id of HelpSupport.
const HelpSupportTypeID = 0x17c6b5f6

func (s *HelpSupport) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.PhoneNumber == "") {
		return false
	}
	if !(s.User == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *HelpSupport) String() string {
	if s == nil {
		return "HelpSupport(nil)"
	}
	type Alias HelpSupport
	return fmt.Sprintf("HelpSupport%+v", Alias(*s))
}

// FillFrom fills HelpSupport from given interface.
func (s *HelpSupport) FillFrom(from interface {
	GetPhoneNumber() (value string)
	GetUser() (value UserClass)
}) {
	s.PhoneNumber = from.GetPhoneNumber()
	s.User = from.GetUser()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpSupport) TypeID() uint32 {
	return HelpSupportTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpSupport) TypeName() string {
	return "help.support"
}

// TypeInfo returns info about TL type.
func (s *HelpSupport) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.support",
		ID:   HelpSupportTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "PhoneNumber",
			SchemaName: "phone_number",
		},
		{
			Name:       "User",
			SchemaName: "user",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *HelpSupport) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.support#17c6b5f6",
		}
	}
	b.PutID(HelpSupportTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *HelpSupport) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.support#17c6b5f6",
		}
	}
	b.PutString(s.PhoneNumber)
	if s.User == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "help.support#17c6b5f6",
			FieldName: "user",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "User",
			},
		}
	}
	if err := s.User.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "help.support#17c6b5f6",
			FieldName:  "user",
			Underlying: err,
		}
	}
	return nil
}

// GetPhoneNumber returns value of PhoneNumber field.
func (s *HelpSupport) GetPhoneNumber() (value string) {
	return s.PhoneNumber
}

// GetUser returns value of User field.
func (s *HelpSupport) GetUser() (value UserClass) {
	return s.User
}

// GetUserAsNotEmpty returns mapped value of User field.
func (s *HelpSupport) GetUserAsNotEmpty() (*User, bool) {
	return s.User.AsNotEmpty()
}

// Decode implements bin.Decoder.
func (s *HelpSupport) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.support#17c6b5f6",
		}
	}
	if err := b.ConsumeID(HelpSupportTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "help.support#17c6b5f6",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *HelpSupport) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.support#17c6b5f6",
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.support#17c6b5f6",
				FieldName:  "phone_number",
				Underlying: err,
			}
		}
		s.PhoneNumber = value
	}
	{
		value, err := DecodeUser(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.support#17c6b5f6",
				FieldName:  "user",
				Underlying: err,
			}
		}
		s.User = value
	}
	return nil
}

// Ensuring interfaces in compile-time for HelpSupport.
var (
	_ bin.Encoder     = &HelpSupport{}
	_ bin.Decoder     = &HelpSupport{}
	_ bin.BareEncoder = &HelpSupport{}
	_ bin.BareDecoder = &HelpSupport{}
)
