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

// PhoneJoinGroupCallPresentationRequest represents TL type `phone.joinGroupCallPresentation#cbea6bc4`.
//
// See https://core.telegram.org/method/phone.joinGroupCallPresentation for reference.
type PhoneJoinGroupCallPresentationRequest struct {
	// Call field of PhoneJoinGroupCallPresentationRequest.
	Call InputGroupCall
	// Params field of PhoneJoinGroupCallPresentationRequest.
	Params DataJSON
}

// PhoneJoinGroupCallPresentationRequestTypeID is TL type id of PhoneJoinGroupCallPresentationRequest.
const PhoneJoinGroupCallPresentationRequestTypeID = 0xcbea6bc4

func (j *PhoneJoinGroupCallPresentationRequest) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Call.Zero()) {
		return false
	}
	if !(j.Params.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *PhoneJoinGroupCallPresentationRequest) String() string {
	if j == nil {
		return "PhoneJoinGroupCallPresentationRequest(nil)"
	}
	type Alias PhoneJoinGroupCallPresentationRequest
	return fmt.Sprintf("PhoneJoinGroupCallPresentationRequest%+v", Alias(*j))
}

// FillFrom fills PhoneJoinGroupCallPresentationRequest from given interface.
func (j *PhoneJoinGroupCallPresentationRequest) FillFrom(from interface {
	GetCall() (value InputGroupCall)
	GetParams() (value DataJSON)
}) {
	j.Call = from.GetCall()
	j.Params = from.GetParams()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PhoneJoinGroupCallPresentationRequest) TypeID() uint32 {
	return PhoneJoinGroupCallPresentationRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PhoneJoinGroupCallPresentationRequest) TypeName() string {
	return "phone.joinGroupCallPresentation"
}

// TypeInfo returns info about TL type.
func (j *PhoneJoinGroupCallPresentationRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "phone.joinGroupCallPresentation",
		ID:   PhoneJoinGroupCallPresentationRequestTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Call",
			SchemaName: "call",
		},
		{
			Name:       "Params",
			SchemaName: "params",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *PhoneJoinGroupCallPresentationRequest) Encode(b *bin.Buffer) error {
	if j == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.joinGroupCallPresentation#cbea6bc4",
		}
	}
	b.PutID(PhoneJoinGroupCallPresentationRequestTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *PhoneJoinGroupCallPresentationRequest) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.joinGroupCallPresentation#cbea6bc4",
		}
	}
	if err := j.Call.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "phone.joinGroupCallPresentation#cbea6bc4",
			FieldName:  "call",
			Underlying: err,
		}
	}
	if err := j.Params.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "phone.joinGroupCallPresentation#cbea6bc4",
			FieldName:  "params",
			Underlying: err,
		}
	}
	return nil
}

// GetCall returns value of Call field.
func (j *PhoneJoinGroupCallPresentationRequest) GetCall() (value InputGroupCall) {
	return j.Call
}

// GetParams returns value of Params field.
func (j *PhoneJoinGroupCallPresentationRequest) GetParams() (value DataJSON) {
	return j.Params
}

// Decode implements bin.Decoder.
func (j *PhoneJoinGroupCallPresentationRequest) Decode(b *bin.Buffer) error {
	if j == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.joinGroupCallPresentation#cbea6bc4",
		}
	}
	if err := b.ConsumeID(PhoneJoinGroupCallPresentationRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "phone.joinGroupCallPresentation#cbea6bc4",
			Underlying: err,
		}
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *PhoneJoinGroupCallPresentationRequest) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.joinGroupCallPresentation#cbea6bc4",
		}
	}
	{
		if err := j.Call.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.joinGroupCallPresentation#cbea6bc4",
				FieldName:  "call",
				Underlying: err,
			}
		}
	}
	{
		if err := j.Params.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.joinGroupCallPresentation#cbea6bc4",
				FieldName:  "params",
				Underlying: err,
			}
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for PhoneJoinGroupCallPresentationRequest.
var (
	_ bin.Encoder     = &PhoneJoinGroupCallPresentationRequest{}
	_ bin.Decoder     = &PhoneJoinGroupCallPresentationRequest{}
	_ bin.BareEncoder = &PhoneJoinGroupCallPresentationRequest{}
	_ bin.BareDecoder = &PhoneJoinGroupCallPresentationRequest{}
)

// PhoneJoinGroupCallPresentation invokes method phone.joinGroupCallPresentation#cbea6bc4 returning error if any.
//
// See https://core.telegram.org/method/phone.joinGroupCallPresentation for reference.
func (c *Client) PhoneJoinGroupCallPresentation(ctx context.Context, request *PhoneJoinGroupCallPresentationRequest) (UpdatesClass, error) {
	var result UpdatesBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.Updates, nil
}
