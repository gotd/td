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

// ContactsGetBlockedRequest represents TL type `contacts.getBlocked#f57c350f`.
// Returns the list of blocked users.
//
// See https://core.telegram.org/method/contacts.getBlocked for reference.
type ContactsGetBlockedRequest struct {
	// The number of list elements to be skipped
	Offset int
	// The number of list elements to be returned
	Limit int
}

// ContactsGetBlockedRequestTypeID is TL type id of ContactsGetBlockedRequest.
const ContactsGetBlockedRequestTypeID = 0xf57c350f

func (g *ContactsGetBlockedRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Offset == 0) {
		return false
	}
	if !(g.Limit == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *ContactsGetBlockedRequest) String() string {
	if g == nil {
		return "ContactsGetBlockedRequest(nil)"
	}
	type Alias ContactsGetBlockedRequest
	return fmt.Sprintf("ContactsGetBlockedRequest%+v", Alias(*g))
}

// FillFrom fills ContactsGetBlockedRequest from given interface.
func (g *ContactsGetBlockedRequest) FillFrom(from interface {
	GetOffset() (value int)
	GetLimit() (value int)
}) {
	g.Offset = from.GetOffset()
	g.Limit = from.GetLimit()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ContactsGetBlockedRequest) TypeID() uint32 {
	return ContactsGetBlockedRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ContactsGetBlockedRequest) TypeName() string {
	return "contacts.getBlocked"
}

// TypeInfo returns info about TL type.
func (g *ContactsGetBlockedRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "contacts.getBlocked",
		ID:   ContactsGetBlockedRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Offset",
			SchemaName: "offset",
		},
		{
			Name:       "Limit",
			SchemaName: "limit",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *ContactsGetBlockedRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "contacts.getBlocked#f57c350f",
		}
	}
	b.PutID(ContactsGetBlockedRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *ContactsGetBlockedRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "contacts.getBlocked#f57c350f",
		}
	}
	b.PutInt(g.Offset)
	b.PutInt(g.Limit)
	return nil
}

// GetOffset returns value of Offset field.
func (g *ContactsGetBlockedRequest) GetOffset() (value int) {
	return g.Offset
}

// GetLimit returns value of Limit field.
func (g *ContactsGetBlockedRequest) GetLimit() (value int) {
	return g.Limit
}

// Decode implements bin.Decoder.
func (g *ContactsGetBlockedRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "contacts.getBlocked#f57c350f",
		}
	}
	if err := b.ConsumeID(ContactsGetBlockedRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "contacts.getBlocked#f57c350f",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *ContactsGetBlockedRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "contacts.getBlocked#f57c350f",
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "contacts.getBlocked#f57c350f",
				FieldName:  "offset",
				Underlying: err,
			}
		}
		g.Offset = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "contacts.getBlocked#f57c350f",
				FieldName:  "limit",
				Underlying: err,
			}
		}
		g.Limit = value
	}
	return nil
}

// Ensuring interfaces in compile-time for ContactsGetBlockedRequest.
var (
	_ bin.Encoder     = &ContactsGetBlockedRequest{}
	_ bin.Decoder     = &ContactsGetBlockedRequest{}
	_ bin.BareEncoder = &ContactsGetBlockedRequest{}
	_ bin.BareDecoder = &ContactsGetBlockedRequest{}
)

// ContactsGetBlocked invokes method contacts.getBlocked#f57c350f returning error if any.
// Returns the list of blocked users.
//
// See https://core.telegram.org/method/contacts.getBlocked for reference.
func (c *Client) ContactsGetBlocked(ctx context.Context, request *ContactsGetBlockedRequest) (ContactsBlockedClass, error) {
	var result ContactsBlockedBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.Blocked, nil
}
