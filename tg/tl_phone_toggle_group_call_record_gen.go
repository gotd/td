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

// PhoneToggleGroupCallRecordRequest represents TL type `phone.toggleGroupCallRecord#c02a66d7`.
//
// See https://core.telegram.org/method/phone.toggleGroupCallRecord for reference.
type PhoneToggleGroupCallRecordRequest struct {
	// Flags field of PhoneToggleGroupCallRecordRequest.
	Flags bin.Fields
	// Start field of PhoneToggleGroupCallRecordRequest.
	Start bool
	// Call field of PhoneToggleGroupCallRecordRequest.
	Call InputGroupCall
	// Title field of PhoneToggleGroupCallRecordRequest.
	//
	// Use SetTitle and GetTitle helpers.
	Title string
}

// PhoneToggleGroupCallRecordRequestTypeID is TL type id of PhoneToggleGroupCallRecordRequest.
const PhoneToggleGroupCallRecordRequestTypeID = 0xc02a66d7

func (t *PhoneToggleGroupCallRecordRequest) Zero() bool {
	if t == nil {
		return true
	}
	if !(t.Flags.Zero()) {
		return false
	}
	if !(t.Start == false) {
		return false
	}
	if !(t.Call.Zero()) {
		return false
	}
	if !(t.Title == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (t *PhoneToggleGroupCallRecordRequest) String() string {
	if t == nil {
		return "PhoneToggleGroupCallRecordRequest(nil)"
	}
	type Alias PhoneToggleGroupCallRecordRequest
	return fmt.Sprintf("PhoneToggleGroupCallRecordRequest%+v", Alias(*t))
}

// FillFrom fills PhoneToggleGroupCallRecordRequest from given interface.
func (t *PhoneToggleGroupCallRecordRequest) FillFrom(from interface {
	GetStart() (value bool)
	GetCall() (value InputGroupCall)
	GetTitle() (value string, ok bool)
}) {
	t.Start = from.GetStart()
	t.Call = from.GetCall()
	if val, ok := from.GetTitle(); ok {
		t.Title = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PhoneToggleGroupCallRecordRequest) TypeID() uint32 {
	return PhoneToggleGroupCallRecordRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PhoneToggleGroupCallRecordRequest) TypeName() string {
	return "phone.toggleGroupCallRecord"
}

// TypeInfo returns info about TL type.
func (t *PhoneToggleGroupCallRecordRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "phone.toggleGroupCallRecord",
		ID:   PhoneToggleGroupCallRecordRequestTypeID,
	}
	if t == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Start",
			SchemaName: "start",
			Null:       !t.Flags.Has(0),
		},
		{
			Name:       "Call",
			SchemaName: "call",
		},
		{
			Name:       "Title",
			SchemaName: "title",
			Null:       !t.Flags.Has(1),
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (t *PhoneToggleGroupCallRecordRequest) Encode(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.toggleGroupCallRecord#c02a66d7",
		}
	}
	b.PutID(PhoneToggleGroupCallRecordRequestTypeID)
	return t.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (t *PhoneToggleGroupCallRecordRequest) EncodeBare(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.toggleGroupCallRecord#c02a66d7",
		}
	}
	if !(t.Start == false) {
		t.Flags.Set(0)
	}
	if !(t.Title == "") {
		t.Flags.Set(1)
	}
	if err := t.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	if err := t.Call.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
			FieldName:  "call",
			Underlying: err,
		}
	}
	if t.Flags.Has(1) {
		b.PutString(t.Title)
	}
	return nil
}

// SetStart sets value of Start conditional field.
func (t *PhoneToggleGroupCallRecordRequest) SetStart(value bool) {
	if value {
		t.Flags.Set(0)
		t.Start = true
	} else {
		t.Flags.Unset(0)
		t.Start = false
	}
}

// GetStart returns value of Start conditional field.
func (t *PhoneToggleGroupCallRecordRequest) GetStart() (value bool) {
	return t.Flags.Has(0)
}

// GetCall returns value of Call field.
func (t *PhoneToggleGroupCallRecordRequest) GetCall() (value InputGroupCall) {
	return t.Call
}

// SetTitle sets value of Title conditional field.
func (t *PhoneToggleGroupCallRecordRequest) SetTitle(value string) {
	t.Flags.Set(1)
	t.Title = value
}

// GetTitle returns value of Title conditional field and
// boolean which is true if field was set.
func (t *PhoneToggleGroupCallRecordRequest) GetTitle() (value string, ok bool) {
	if !t.Flags.Has(1) {
		return value, false
	}
	return t.Title, true
}

// Decode implements bin.Decoder.
func (t *PhoneToggleGroupCallRecordRequest) Decode(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.toggleGroupCallRecord#c02a66d7",
		}
	}
	if err := b.ConsumeID(PhoneToggleGroupCallRecordRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
			Underlying: err,
		}
	}
	return t.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (t *PhoneToggleGroupCallRecordRequest) DecodeBare(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.toggleGroupCallRecord#c02a66d7",
		}
	}
	{
		if err := t.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	t.Start = t.Flags.Has(0)
	{
		if err := t.Call.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
				FieldName:  "call",
				Underlying: err,
			}
		}
	}
	if t.Flags.Has(1) {
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.toggleGroupCallRecord#c02a66d7",
				FieldName:  "title",
				Underlying: err,
			}
		}
		t.Title = value
	}
	return nil
}

// Ensuring interfaces in compile-time for PhoneToggleGroupCallRecordRequest.
var (
	_ bin.Encoder     = &PhoneToggleGroupCallRecordRequest{}
	_ bin.Decoder     = &PhoneToggleGroupCallRecordRequest{}
	_ bin.BareEncoder = &PhoneToggleGroupCallRecordRequest{}
	_ bin.BareDecoder = &PhoneToggleGroupCallRecordRequest{}
)

// PhoneToggleGroupCallRecord invokes method phone.toggleGroupCallRecord#c02a66d7 returning error if any.
//
// See https://core.telegram.org/method/phone.toggleGroupCallRecord for reference.
func (c *Client) PhoneToggleGroupCallRecord(ctx context.Context, request *PhoneToggleGroupCallRecordRequest) (UpdatesClass, error) {
	var result UpdatesBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.Updates, nil
}
