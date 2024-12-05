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

// FoundAffiliatePrograms represents TL type `foundAffiliatePrograms#b6228108`.
type FoundAffiliatePrograms struct {
	// The total number of found affiliate programs
	TotalCount int32
	// The list of affiliate programs
	Programs []FoundAffiliateProgram
	// The offset for the next request. If empty, then there are no more results
	NextOffset string
}

// FoundAffiliateProgramsTypeID is TL type id of FoundAffiliatePrograms.
const FoundAffiliateProgramsTypeID = 0xb6228108

// Ensuring interfaces in compile-time for FoundAffiliatePrograms.
var (
	_ bin.Encoder     = &FoundAffiliatePrograms{}
	_ bin.Decoder     = &FoundAffiliatePrograms{}
	_ bin.BareEncoder = &FoundAffiliatePrograms{}
	_ bin.BareDecoder = &FoundAffiliatePrograms{}
)

func (f *FoundAffiliatePrograms) Zero() bool {
	if f == nil {
		return true
	}
	if !(f.TotalCount == 0) {
		return false
	}
	if !(f.Programs == nil) {
		return false
	}
	if !(f.NextOffset == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (f *FoundAffiliatePrograms) String() string {
	if f == nil {
		return "FoundAffiliatePrograms(nil)"
	}
	type Alias FoundAffiliatePrograms
	return fmt.Sprintf("FoundAffiliatePrograms%+v", Alias(*f))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*FoundAffiliatePrograms) TypeID() uint32 {
	return FoundAffiliateProgramsTypeID
}

// TypeName returns name of type in TL schema.
func (*FoundAffiliatePrograms) TypeName() string {
	return "foundAffiliatePrograms"
}

// TypeInfo returns info about TL type.
func (f *FoundAffiliatePrograms) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "foundAffiliatePrograms",
		ID:   FoundAffiliateProgramsTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "TotalCount",
			SchemaName: "total_count",
		},
		{
			Name:       "Programs",
			SchemaName: "programs",
		},
		{
			Name:       "NextOffset",
			SchemaName: "next_offset",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (f *FoundAffiliatePrograms) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode foundAffiliatePrograms#b6228108 as nil")
	}
	b.PutID(FoundAffiliateProgramsTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *FoundAffiliatePrograms) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode foundAffiliatePrograms#b6228108 as nil")
	}
	b.PutInt32(f.TotalCount)
	b.PutInt(len(f.Programs))
	for idx, v := range f.Programs {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare foundAffiliatePrograms#b6228108: field programs element with index %d: %w", idx, err)
		}
	}
	b.PutString(f.NextOffset)
	return nil
}

// Decode implements bin.Decoder.
func (f *FoundAffiliatePrograms) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode foundAffiliatePrograms#b6228108 to nil")
	}
	if err := b.ConsumeID(FoundAffiliateProgramsTypeID); err != nil {
		return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *FoundAffiliatePrograms) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode foundAffiliatePrograms#b6228108 to nil")
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field total_count: %w", err)
		}
		f.TotalCount = value
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field programs: %w", err)
		}

		if headerLen > 0 {
			f.Programs = make([]FoundAffiliateProgram, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value FoundAffiliateProgram
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare foundAffiliatePrograms#b6228108: field programs: %w", err)
			}
			f.Programs = append(f.Programs, value)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field next_offset: %w", err)
		}
		f.NextOffset = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (f *FoundAffiliatePrograms) EncodeTDLibJSON(b tdjson.Encoder) error {
	if f == nil {
		return fmt.Errorf("can't encode foundAffiliatePrograms#b6228108 as nil")
	}
	b.ObjStart()
	b.PutID("foundAffiliatePrograms")
	b.Comma()
	b.FieldStart("total_count")
	b.PutInt32(f.TotalCount)
	b.Comma()
	b.FieldStart("programs")
	b.ArrStart()
	for idx, v := range f.Programs {
		if err := v.EncodeTDLibJSON(b); err != nil {
			return fmt.Errorf("unable to encode foundAffiliatePrograms#b6228108: field programs element with index %d: %w", idx, err)
		}
		b.Comma()
	}
	b.StripComma()
	b.ArrEnd()
	b.Comma()
	b.FieldStart("next_offset")
	b.PutString(f.NextOffset)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (f *FoundAffiliatePrograms) DecodeTDLibJSON(b tdjson.Decoder) error {
	if f == nil {
		return fmt.Errorf("can't decode foundAffiliatePrograms#b6228108 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("foundAffiliatePrograms"); err != nil {
				return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: %w", err)
			}
		case "total_count":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field total_count: %w", err)
			}
			f.TotalCount = value
		case "programs":
			if err := b.Arr(func(b tdjson.Decoder) error {
				var value FoundAffiliateProgram
				if err := value.DecodeTDLibJSON(b); err != nil {
					return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field programs: %w", err)
				}
				f.Programs = append(f.Programs, value)
				return nil
			}); err != nil {
				return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field programs: %w", err)
			}
		case "next_offset":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode foundAffiliatePrograms#b6228108: field next_offset: %w", err)
			}
			f.NextOffset = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetTotalCount returns value of TotalCount field.
func (f *FoundAffiliatePrograms) GetTotalCount() (value int32) {
	if f == nil {
		return
	}
	return f.TotalCount
}

// GetPrograms returns value of Programs field.
func (f *FoundAffiliatePrograms) GetPrograms() (value []FoundAffiliateProgram) {
	if f == nil {
		return
	}
	return f.Programs
}

// GetNextOffset returns value of NextOffset field.
func (f *FoundAffiliatePrograms) GetNextOffset() (value string) {
	if f == nil {
		return
	}
	return f.NextOffset
}