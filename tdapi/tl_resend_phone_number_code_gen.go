// Code generated by gotdgen, DO NOT EDIT.

package tdapi

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdjson"
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
	_ = tdjson.Encoder{}
)

// ResendPhoneNumberCodeRequest represents TL type `resendPhoneNumberCode#2dc1f7c8`.
type ResendPhoneNumberCodeRequest struct {
}

// ResendPhoneNumberCodeRequestTypeID is TL type id of ResendPhoneNumberCodeRequest.
const ResendPhoneNumberCodeRequestTypeID = 0x2dc1f7c8

// Ensuring interfaces in compile-time for ResendPhoneNumberCodeRequest.
var (
	_ bin.Encoder     = &ResendPhoneNumberCodeRequest{}
	_ bin.Decoder     = &ResendPhoneNumberCodeRequest{}
	_ bin.BareEncoder = &ResendPhoneNumberCodeRequest{}
	_ bin.BareDecoder = &ResendPhoneNumberCodeRequest{}
)

func (r *ResendPhoneNumberCodeRequest) Zero() bool {
	if r == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (r *ResendPhoneNumberCodeRequest) String() string {
	if r == nil {
		return "ResendPhoneNumberCodeRequest(nil)"
	}
	type Alias ResendPhoneNumberCodeRequest
	return fmt.Sprintf("ResendPhoneNumberCodeRequest%+v", Alias(*r))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ResendPhoneNumberCodeRequest) TypeID() uint32 {
	return ResendPhoneNumberCodeRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ResendPhoneNumberCodeRequest) TypeName() string {
	return "resendPhoneNumberCode"
}

// TypeInfo returns info about TL type.
func (r *ResendPhoneNumberCodeRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "resendPhoneNumberCode",
		ID:   ResendPhoneNumberCodeRequestTypeID,
	}
	if r == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (r *ResendPhoneNumberCodeRequest) Encode(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't encode resendPhoneNumberCode#2dc1f7c8 as nil")
	}
	b.PutID(ResendPhoneNumberCodeRequestTypeID)
	return r.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (r *ResendPhoneNumberCodeRequest) EncodeBare(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't encode resendPhoneNumberCode#2dc1f7c8 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (r *ResendPhoneNumberCodeRequest) Decode(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't decode resendPhoneNumberCode#2dc1f7c8 to nil")
	}
	if err := b.ConsumeID(ResendPhoneNumberCodeRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode resendPhoneNumberCode#2dc1f7c8: %w", err)
	}
	return r.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (r *ResendPhoneNumberCodeRequest) DecodeBare(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't decode resendPhoneNumberCode#2dc1f7c8 to nil")
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (r *ResendPhoneNumberCodeRequest) EncodeTDLibJSON(b tdjson.Encoder) error {
	if r == nil {
		return fmt.Errorf("can't encode resendPhoneNumberCode#2dc1f7c8 as nil")
	}
	b.ObjStart()
	b.PutID("resendPhoneNumberCode")
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (r *ResendPhoneNumberCodeRequest) DecodeTDLibJSON(b tdjson.Decoder) error {
	if r == nil {
		return fmt.Errorf("can't decode resendPhoneNumberCode#2dc1f7c8 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("resendPhoneNumberCode"); err != nil {
				return fmt.Errorf("unable to decode resendPhoneNumberCode#2dc1f7c8: %w", err)
			}
		default:
			return b.Skip()
		}
		return nil
	})
}

// ResendPhoneNumberCode invokes method resendPhoneNumberCode#2dc1f7c8 returning error if any.
func (c *Client) ResendPhoneNumberCode(ctx context.Context) (*AuthenticationCodeInfo, error) {
	var result AuthenticationCodeInfo

	request := &ResendPhoneNumberCodeRequest{}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}