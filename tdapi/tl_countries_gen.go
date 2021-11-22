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

// Countries represents TL type `countries#94b50e0f`.
type Countries struct {
	// The list of countries
	Countries []CountryInfo
}

// CountriesTypeID is TL type id of Countries.
const CountriesTypeID = 0x94b50e0f

// Ensuring interfaces in compile-time for Countries.
var (
	_ bin.Encoder     = &Countries{}
	_ bin.Decoder     = &Countries{}
	_ bin.BareEncoder = &Countries{}
	_ bin.BareDecoder = &Countries{}
)

func (c *Countries) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.Countries == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *Countries) String() string {
	if c == nil {
		return "Countries(nil)"
	}
	type Alias Countries
	return fmt.Sprintf("Countries%+v", Alias(*c))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*Countries) TypeID() uint32 {
	return CountriesTypeID
}

// TypeName returns name of type in TL schema.
func (*Countries) TypeName() string {
	return "countries"
}

// TypeInfo returns info about TL type.
func (c *Countries) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "countries",
		ID:   CountriesTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Countries",
			SchemaName: "countries",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *Countries) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode countries#94b50e0f as nil")
	}
	b.PutID(CountriesTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *Countries) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode countries#94b50e0f as nil")
	}
	b.PutInt(len(c.Countries))
	for idx, v := range c.Countries {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare countries#94b50e0f: field countries element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (c *Countries) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode countries#94b50e0f to nil")
	}
	if err := b.ConsumeID(CountriesTypeID); err != nil {
		return fmt.Errorf("unable to decode countries#94b50e0f: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *Countries) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode countries#94b50e0f to nil")
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode countries#94b50e0f: field countries: %w", err)
		}

		if headerLen > 0 {
			c.Countries = make([]CountryInfo, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value CountryInfo
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare countries#94b50e0f: field countries: %w", err)
			}
			c.Countries = append(c.Countries, value)
		}
	}
	return nil
}

// GetCountries returns value of Countries field.
func (c *Countries) GetCountries() (value []CountryInfo) {
	return c.Countries
}