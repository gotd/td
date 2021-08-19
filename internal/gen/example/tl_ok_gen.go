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

// Ok represents TL type `ok#d4edbe69`.
//
// See https://localhost:80/doc/constructor/ok for reference.
type Ok struct {
}

// OkTypeID is TL type id of Ok.
const OkTypeID = 0xd4edbe69

func (o *Ok) Zero() bool {
	if o == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (o *Ok) String() string {
	if o == nil {
		return "Ok(nil)"
	}
	type Alias Ok
	return fmt.Sprintf("Ok%+v", Alias(*o))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*Ok) TypeID() uint32 {
	return OkTypeID
}

// TypeName returns name of type in TL schema.
func (*Ok) TypeName() string {
	return "ok"
}

// TypeInfo returns info about TL type.
func (o *Ok) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "ok",
		ID:   OkTypeID,
	}
	if o == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (o *Ok) Encode(b *bin.Buffer) error {
	if o == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "ok#d4edbe69",
		}
	}
	b.PutID(OkTypeID)
	return o.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (o *Ok) EncodeBare(b *bin.Buffer) error {
	if o == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "ok#d4edbe69",
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (o *Ok) Decode(b *bin.Buffer) error {
	if o == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "ok#d4edbe69",
		}
	}
	if err := b.ConsumeID(OkTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "ok#d4edbe69",
			Underlying: err,
		}
	}
	return o.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (o *Ok) DecodeBare(b *bin.Buffer) error {
	if o == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "ok#d4edbe69",
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for Ok.
var (
	_ bin.Encoder     = &Ok{}
	_ bin.Decoder     = &Ok{}
	_ bin.BareEncoder = &Ok{}
	_ bin.BareDecoder = &Ok{}
)
