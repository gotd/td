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

// TestInt represents TL type `testInt#ddbd2c09`.
//
// See https://localhost:80/doc/constructor/testInt for reference.
type TestInt struct {
	// Number
	Value int32
}

// TestIntTypeID is TL type id of TestInt.
const TestIntTypeID = 0xddbd2c09

func (t *TestInt) Zero() bool {
	if t == nil {
		return true
	}
	if !(t.Value == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (t *TestInt) String() string {
	if t == nil {
		return "TestInt(nil)"
	}
	type Alias TestInt
	return fmt.Sprintf("TestInt%+v", Alias(*t))
}

// FillFrom fills TestInt from given interface.
func (t *TestInt) FillFrom(from interface {
	GetValue() (value int32)
}) {
	t.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*TestInt) TypeID() uint32 {
	return TestIntTypeID
}

// TypeName returns name of type in TL schema.
func (*TestInt) TypeName() string {
	return "testInt"
}

// TypeInfo returns info about TL type.
func (t *TestInt) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "testInt",
		ID:   TestIntTypeID,
	}
	if t == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (t *TestInt) Encode(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "testInt#ddbd2c09",
		}
	}
	b.PutID(TestIntTypeID)
	return t.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (t *TestInt) EncodeBare(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "testInt#ddbd2c09",
		}
	}
	b.PutInt32(t.Value)
	return nil
}

// GetValue returns value of Value field.
func (t *TestInt) GetValue() (value int32) {
	return t.Value
}

// Decode implements bin.Decoder.
func (t *TestInt) Decode(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "testInt#ddbd2c09",
		}
	}
	if err := b.ConsumeID(TestIntTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "testInt#ddbd2c09",
			Underlying: err,
		}
	}
	return t.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (t *TestInt) DecodeBare(b *bin.Buffer) error {
	if t == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "testInt#ddbd2c09",
		}
	}
	{
		value, err := b.Int32()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "testInt#ddbd2c09",
				FieldName:  "value",
				Underlying: err,
			}
		}
		t.Value = value
	}
	return nil
}

// Ensuring interfaces in compile-time for TestInt.
var (
	_ bin.Encoder     = &TestInt{}
	_ bin.Decoder     = &TestInt{}
	_ bin.BareEncoder = &TestInt{}
	_ bin.BareDecoder = &TestInt{}
)
