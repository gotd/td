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

// HelpCountryCode represents TL type `help.countryCode#4203c5ef`.
// Country code and phone number pattern of a specific country
//
// See https://core.telegram.org/constructor/help.countryCode for reference.
type HelpCountryCode struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// ISO country code
	CountryCode string
	// Possible phone prefixes
	//
	// Use SetPrefixes and GetPrefixes helpers.
	Prefixes []string
	// Phone patterns: for example, XXX XXX XXX
	//
	// Use SetPatterns and GetPatterns helpers.
	Patterns []string
}

// HelpCountryCodeTypeID is TL type id of HelpCountryCode.
const HelpCountryCodeTypeID = 0x4203c5ef

func (c *HelpCountryCode) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.Flags.Zero()) {
		return false
	}
	if !(c.CountryCode == "") {
		return false
	}
	if !(c.Prefixes == nil) {
		return false
	}
	if !(c.Patterns == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *HelpCountryCode) String() string {
	if c == nil {
		return "HelpCountryCode(nil)"
	}
	type Alias HelpCountryCode
	return fmt.Sprintf("HelpCountryCode%+v", Alias(*c))
}

// FillFrom fills HelpCountryCode from given interface.
func (c *HelpCountryCode) FillFrom(from interface {
	GetCountryCode() (value string)
	GetPrefixes() (value []string, ok bool)
	GetPatterns() (value []string, ok bool)
}) {
	c.CountryCode = from.GetCountryCode()
	if val, ok := from.GetPrefixes(); ok {
		c.Prefixes = val
	}

	if val, ok := from.GetPatterns(); ok {
		c.Patterns = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpCountryCode) TypeID() uint32 {
	return HelpCountryCodeTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpCountryCode) TypeName() string {
	return "help.countryCode"
}

// TypeInfo returns info about TL type.
func (c *HelpCountryCode) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.countryCode",
		ID:   HelpCountryCodeTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "CountryCode",
			SchemaName: "country_code",
		},
		{
			Name:       "Prefixes",
			SchemaName: "prefixes",
			Null:       !c.Flags.Has(0),
		},
		{
			Name:       "Patterns",
			SchemaName: "patterns",
			Null:       !c.Flags.Has(1),
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *HelpCountryCode) Encode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.countryCode#4203c5ef",
		}
	}
	b.PutID(HelpCountryCodeTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *HelpCountryCode) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.countryCode#4203c5ef",
		}
	}
	if !(c.Prefixes == nil) {
		c.Flags.Set(0)
	}
	if !(c.Patterns == nil) {
		c.Flags.Set(1)
	}
	if err := c.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "help.countryCode#4203c5ef",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	b.PutString(c.CountryCode)
	if c.Flags.Has(0) {
		b.PutVectorHeader(len(c.Prefixes))
		for _, v := range c.Prefixes {
			b.PutString(v)
		}
	}
	if c.Flags.Has(1) {
		b.PutVectorHeader(len(c.Patterns))
		for _, v := range c.Patterns {
			b.PutString(v)
		}
	}
	return nil
}

// GetCountryCode returns value of CountryCode field.
func (c *HelpCountryCode) GetCountryCode() (value string) {
	return c.CountryCode
}

// SetPrefixes sets value of Prefixes conditional field.
func (c *HelpCountryCode) SetPrefixes(value []string) {
	c.Flags.Set(0)
	c.Prefixes = value
}

// GetPrefixes returns value of Prefixes conditional field and
// boolean which is true if field was set.
func (c *HelpCountryCode) GetPrefixes() (value []string, ok bool) {
	if !c.Flags.Has(0) {
		return value, false
	}
	return c.Prefixes, true
}

// SetPatterns sets value of Patterns conditional field.
func (c *HelpCountryCode) SetPatterns(value []string) {
	c.Flags.Set(1)
	c.Patterns = value
}

// GetPatterns returns value of Patterns conditional field and
// boolean which is true if field was set.
func (c *HelpCountryCode) GetPatterns() (value []string, ok bool) {
	if !c.Flags.Has(1) {
		return value, false
	}
	return c.Patterns, true
}

// Decode implements bin.Decoder.
func (c *HelpCountryCode) Decode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.countryCode#4203c5ef",
		}
	}
	if err := b.ConsumeID(HelpCountryCodeTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "help.countryCode#4203c5ef",
			Underlying: err,
		}
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *HelpCountryCode) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.countryCode#4203c5ef",
		}
	}
	{
		if err := c.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.countryCode#4203c5ef",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.countryCode#4203c5ef",
				FieldName:  "country_code",
				Underlying: err,
			}
		}
		c.CountryCode = value
	}
	if c.Flags.Has(0) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.countryCode#4203c5ef",
				FieldName:  "prefixes",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			c.Prefixes = make([]string, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.String()
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "help.countryCode#4203c5ef",
					FieldName:  "prefixes",
					Underlying: err,
				}
			}
			c.Prefixes = append(c.Prefixes, value)
		}
	}
	if c.Flags.Has(1) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "help.countryCode#4203c5ef",
				FieldName:  "patterns",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			c.Patterns = make([]string, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.String()
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "help.countryCode#4203c5ef",
					FieldName:  "patterns",
					Underlying: err,
				}
			}
			c.Patterns = append(c.Patterns, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for HelpCountryCode.
var (
	_ bin.Encoder     = &HelpCountryCode{}
	_ bin.Decoder     = &HelpCountryCode{}
	_ bin.BareEncoder = &HelpCountryCode{}
	_ bin.BareDecoder = &HelpCountryCode{}
)
