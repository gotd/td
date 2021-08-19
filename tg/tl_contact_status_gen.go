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

// ContactStatus represents TL type `contactStatus#d3680c61`.
// Contact status: online / offline.
//
// See https://core.telegram.org/constructor/contactStatus for reference.
type ContactStatus struct {
	// User identifier
	UserID int
	// Online status
	Status UserStatusClass
}

// ContactStatusTypeID is TL type id of ContactStatus.
const ContactStatusTypeID = 0xd3680c61

func (c *ContactStatus) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.UserID == 0) {
		return false
	}
	if !(c.Status == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *ContactStatus) String() string {
	if c == nil {
		return "ContactStatus(nil)"
	}
	type Alias ContactStatus
	return fmt.Sprintf("ContactStatus%+v", Alias(*c))
}

// FillFrom fills ContactStatus from given interface.
func (c *ContactStatus) FillFrom(from interface {
	GetUserID() (value int)
	GetStatus() (value UserStatusClass)
}) {
	c.UserID = from.GetUserID()
	c.Status = from.GetStatus()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ContactStatus) TypeID() uint32 {
	return ContactStatusTypeID
}

// TypeName returns name of type in TL schema.
func (*ContactStatus) TypeName() string {
	return "contactStatus"
}

// TypeInfo returns info about TL type.
func (c *ContactStatus) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "contactStatus",
		ID:   ContactStatusTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
		{
			Name:       "Status",
			SchemaName: "status",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *ContactStatus) Encode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "contactStatus#d3680c61",
		}
	}
	b.PutID(ContactStatusTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *ContactStatus) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "contactStatus#d3680c61",
		}
	}
	b.PutInt(c.UserID)
	if c.Status == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "contactStatus#d3680c61",
			FieldName: "status",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "UserStatus",
			},
		}
	}
	if err := c.Status.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "contactStatus#d3680c61",
			FieldName:  "status",
			Underlying: err,
		}
	}
	return nil
}

// GetUserID returns value of UserID field.
func (c *ContactStatus) GetUserID() (value int) {
	return c.UserID
}

// GetStatus returns value of Status field.
func (c *ContactStatus) GetStatus() (value UserStatusClass) {
	return c.Status
}

// Decode implements bin.Decoder.
func (c *ContactStatus) Decode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "contactStatus#d3680c61",
		}
	}
	if err := b.ConsumeID(ContactStatusTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "contactStatus#d3680c61",
			Underlying: err,
		}
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *ContactStatus) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "contactStatus#d3680c61",
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "contactStatus#d3680c61",
				FieldName:  "user_id",
				Underlying: err,
			}
		}
		c.UserID = value
	}
	{
		value, err := DecodeUserStatus(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "contactStatus#d3680c61",
				FieldName:  "status",
				Underlying: err,
			}
		}
		c.Status = value
	}
	return nil
}

// Ensuring interfaces in compile-time for ContactStatus.
var (
	_ bin.Encoder     = &ContactStatus{}
	_ bin.Decoder     = &ContactStatus{}
	_ bin.BareEncoder = &ContactStatus{}
	_ bin.BareDecoder = &ContactStatus{}
)
