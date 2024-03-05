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

// False represents TL type `false#bc799737`.
//
// See https://localhost:80/doc/constructor/false for reference.
type False struct {
}

// FalseTypeID is TL type id of False.
const FalseTypeID = 0xbc799737

// construct implements constructor of BoolClass.
func (f False) construct() BoolClass { return &f }

// Ensuring interfaces in compile-time for False.
var (
	_ bin.Encoder     = &False{}
	_ bin.Decoder     = &False{}
	_ bin.BareEncoder = &False{}
	_ bin.BareDecoder = &False{}

	_ BoolClass = &False{}
)

func (f *False) Zero() bool {
	if f == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (f *False) String() string {
	if f == nil {
		return "False(nil)"
	}
	type Alias False
	return fmt.Sprintf("False%+v", Alias(*f))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*False) TypeID() uint32 {
	return FalseTypeID
}

// TypeName returns name of type in TL schema.
func (*False) TypeName() string {
	return "false"
}

// TypeInfo returns info about TL type.
func (f *False) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "false",
		ID:   FalseTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (f *False) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode false#bc799737 as nil")
	}
	b.PutID(FalseTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *False) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode false#bc799737 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (f *False) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode false#bc799737 to nil")
	}
	if err := b.ConsumeID(FalseTypeID); err != nil {
		return fmt.Errorf("unable to decode false#bc799737: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *False) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode false#bc799737 to nil")
	}
	return nil
}

// True represents TL type `true#997275b5`.
//
// See https://localhost:80/doc/constructor/true for reference.
type True struct {
}

// TrueTypeID is TL type id of True.
const TrueTypeID = 0x997275b5

// construct implements constructor of BoolClass.
func (t True) construct() BoolClass { return &t }

// Ensuring interfaces in compile-time for True.
var (
	_ bin.Encoder     = &True{}
	_ bin.Decoder     = &True{}
	_ bin.BareEncoder = &True{}
	_ bin.BareDecoder = &True{}

	_ BoolClass = &True{}
)

func (t *True) Zero() bool {
	if t == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (t *True) String() string {
	if t == nil {
		return "True(nil)"
	}
	type Alias True
	return fmt.Sprintf("True%+v", Alias(*t))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*True) TypeID() uint32 {
	return TrueTypeID
}

// TypeName returns name of type in TL schema.
func (*True) TypeName() string {
	return "true"
}

// TypeInfo returns info about TL type.
func (t *True) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "true",
		ID:   TrueTypeID,
	}
	if t == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (t *True) Encode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode true#997275b5 as nil")
	}
	b.PutID(TrueTypeID)
	return t.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (t *True) EncodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode true#997275b5 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (t *True) Decode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode true#997275b5 to nil")
	}
	if err := b.ConsumeID(TrueTypeID); err != nil {
		return fmt.Errorf("unable to decode true#997275b5: %w", err)
	}
	return t.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (t *True) DecodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode true#997275b5 to nil")
	}
	return nil
}

// BoolClassName is schema name of BoolClass.
const BoolClassName = "Bool"

// BoolClass represents Bool generic type.
//
// See https://localhost:80/doc/type/Bool for reference.
//
// Example:
//
//	g, err := td.DecodeBool(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *td.False: // false#bc799737
//	case *td.True: // true#997275b5
//	default: panic(v)
//	}
type BoolClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() BoolClass

	// TypeID returns type id in TL schema.
	//
	// See https://core.telegram.org/mtproto/TL-tl#remarks.
	TypeID() uint32
	// TypeName returns name of type in TL schema.
	TypeName() string
	// String implements fmt.Stringer.
	String() string
	// Zero returns true if current object has a zero value.
	Zero() bool
}

// DecodeBool implements binary de-serialization for BoolClass.
func DecodeBool(buf *bin.Buffer) (BoolClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case FalseTypeID:
		// Decoding false#bc799737.
		v := False{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode BoolClass: %w", err)
		}
		return &v, nil
	case TrueTypeID:
		// Decoding true#997275b5.
		v := True{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode BoolClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode BoolClass: %w", bin.NewUnexpectedID(id))
	}
}

// Bool boxes the BoolClass providing a helper.
type BoolBox struct {
	Bool BoolClass
}

// Decode implements bin.Decoder for BoolBox.
func (b *BoolBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode BoolBox to nil")
	}
	v, err := DecodeBool(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.Bool = v
	return nil
}

// Encode implements bin.Encode for BoolBox.
func (b *BoolBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.Bool == nil {
		return fmt.Errorf("unable to encode BoolClass as nil")
	}
	return b.Bool.Encode(buf)
}