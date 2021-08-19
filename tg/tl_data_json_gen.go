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

// DataJSON represents TL type `dataJSON#7d748d04`.
// Represents a json-encoded object
//
// See https://core.telegram.org/constructor/dataJSON for reference.
type DataJSON struct {
	// JSON-encoded object
	Data string
}

// DataJSONTypeID is TL type id of DataJSON.
const DataJSONTypeID = 0x7d748d04

func (d *DataJSON) Zero() bool {
	if d == nil {
		return true
	}
	if !(d.Data == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (d *DataJSON) String() string {
	if d == nil {
		return "DataJSON(nil)"
	}
	type Alias DataJSON
	return fmt.Sprintf("DataJSON%+v", Alias(*d))
}

// FillFrom fills DataJSON from given interface.
func (d *DataJSON) FillFrom(from interface {
	GetData() (value string)
}) {
	d.Data = from.GetData()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*DataJSON) TypeID() uint32 {
	return DataJSONTypeID
}

// TypeName returns name of type in TL schema.
func (*DataJSON) TypeName() string {
	return "dataJSON"
}

// TypeInfo returns info about TL type.
func (d *DataJSON) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "dataJSON",
		ID:   DataJSONTypeID,
	}
	if d == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Data",
			SchemaName: "data",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (d *DataJSON) Encode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dataJSON#7d748d04",
		}
	}
	b.PutID(DataJSONTypeID)
	return d.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (d *DataJSON) EncodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dataJSON#7d748d04",
		}
	}
	b.PutString(d.Data)
	return nil
}

// GetData returns value of Data field.
func (d *DataJSON) GetData() (value string) {
	return d.Data
}

// Decode implements bin.Decoder.
func (d *DataJSON) Decode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dataJSON#7d748d04",
		}
	}
	if err := b.ConsumeID(DataJSONTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "dataJSON#7d748d04",
			Underlying: err,
		}
	}
	return d.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (d *DataJSON) DecodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dataJSON#7d748d04",
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dataJSON#7d748d04",
				FieldName:  "data",
				Underlying: err,
			}
		}
		d.Data = value
	}
	return nil
}

// Ensuring interfaces in compile-time for DataJSON.
var (
	_ bin.Encoder     = &DataJSON{}
	_ bin.Decoder     = &DataJSON{}
	_ bin.BareEncoder = &DataJSON{}
	_ bin.BareDecoder = &DataJSON{}
)
