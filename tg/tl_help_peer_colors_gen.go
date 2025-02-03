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

// HelpPeerColorsNotModified represents TL type `help.peerColorsNotModified#2ba1f5ce`.
// The list of color palettes has not changed.
//
// See https://core.telegram.org/constructor/help.peerColorsNotModified for reference.
type HelpPeerColorsNotModified struct {
}

// HelpPeerColorsNotModifiedTypeID is TL type id of HelpPeerColorsNotModified.
const HelpPeerColorsNotModifiedTypeID = 0x2ba1f5ce

// construct implements constructor of HelpPeerColorsClass.
func (p HelpPeerColorsNotModified) construct() HelpPeerColorsClass { return &p }

// Ensuring interfaces in compile-time for HelpPeerColorsNotModified.
var (
	_ bin.Encoder     = &HelpPeerColorsNotModified{}
	_ bin.Decoder     = &HelpPeerColorsNotModified{}
	_ bin.BareEncoder = &HelpPeerColorsNotModified{}
	_ bin.BareDecoder = &HelpPeerColorsNotModified{}

	_ HelpPeerColorsClass = &HelpPeerColorsNotModified{}
)

func (p *HelpPeerColorsNotModified) Zero() bool {
	if p == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (p *HelpPeerColorsNotModified) String() string {
	if p == nil {
		return "HelpPeerColorsNotModified(nil)"
	}
	type Alias HelpPeerColorsNotModified
	return fmt.Sprintf("HelpPeerColorsNotModified%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpPeerColorsNotModified) TypeID() uint32 {
	return HelpPeerColorsNotModifiedTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpPeerColorsNotModified) TypeName() string {
	return "help.peerColorsNotModified"
}

// TypeInfo returns info about TL type.
func (p *HelpPeerColorsNotModified) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.peerColorsNotModified",
		ID:   HelpPeerColorsNotModifiedTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (p *HelpPeerColorsNotModified) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.peerColorsNotModified#2ba1f5ce as nil")
	}
	b.PutID(HelpPeerColorsNotModifiedTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *HelpPeerColorsNotModified) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.peerColorsNotModified#2ba1f5ce as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *HelpPeerColorsNotModified) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.peerColorsNotModified#2ba1f5ce to nil")
	}
	if err := b.ConsumeID(HelpPeerColorsNotModifiedTypeID); err != nil {
		return fmt.Errorf("unable to decode help.peerColorsNotModified#2ba1f5ce: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *HelpPeerColorsNotModified) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.peerColorsNotModified#2ba1f5ce to nil")
	}
	return nil
}

// HelpPeerColors represents TL type `help.peerColors#f8ed08`.
// Contains info about multiple color palettes »¹.
//
// Links:
//  1. https://core.telegram.org/api/colors
//
// See https://core.telegram.org/constructor/help.peerColors for reference.
type HelpPeerColors struct {
	// Hash used for caching, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets#hash-generation
	Hash int
	// Usable color palettes¹.
	//
	// Links:
	//  1) https://core.telegram.org/api/colors
	Colors []HelpPeerColorOption
}

// HelpPeerColorsTypeID is TL type id of HelpPeerColors.
const HelpPeerColorsTypeID = 0xf8ed08

// construct implements constructor of HelpPeerColorsClass.
func (p HelpPeerColors) construct() HelpPeerColorsClass { return &p }

// Ensuring interfaces in compile-time for HelpPeerColors.
var (
	_ bin.Encoder     = &HelpPeerColors{}
	_ bin.Decoder     = &HelpPeerColors{}
	_ bin.BareEncoder = &HelpPeerColors{}
	_ bin.BareDecoder = &HelpPeerColors{}

	_ HelpPeerColorsClass = &HelpPeerColors{}
)

func (p *HelpPeerColors) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.Hash == 0) {
		return false
	}
	if !(p.Colors == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *HelpPeerColors) String() string {
	if p == nil {
		return "HelpPeerColors(nil)"
	}
	type Alias HelpPeerColors
	return fmt.Sprintf("HelpPeerColors%+v", Alias(*p))
}

// FillFrom fills HelpPeerColors from given interface.
func (p *HelpPeerColors) FillFrom(from interface {
	GetHash() (value int)
	GetColors() (value []HelpPeerColorOption)
}) {
	p.Hash = from.GetHash()
	p.Colors = from.GetColors()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpPeerColors) TypeID() uint32 {
	return HelpPeerColorsTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpPeerColors) TypeName() string {
	return "help.peerColors"
}

// TypeInfo returns info about TL type.
func (p *HelpPeerColors) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.peerColors",
		ID:   HelpPeerColorsTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Hash",
			SchemaName: "hash",
		},
		{
			Name:       "Colors",
			SchemaName: "colors",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *HelpPeerColors) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.peerColors#f8ed08 as nil")
	}
	b.PutID(HelpPeerColorsTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *HelpPeerColors) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.peerColors#f8ed08 as nil")
	}
	b.PutInt(p.Hash)
	b.PutVectorHeader(len(p.Colors))
	for idx, v := range p.Colors {
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode help.peerColors#f8ed08: field colors element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *HelpPeerColors) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.peerColors#f8ed08 to nil")
	}
	if err := b.ConsumeID(HelpPeerColorsTypeID); err != nil {
		return fmt.Errorf("unable to decode help.peerColors#f8ed08: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *HelpPeerColors) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.peerColors#f8ed08 to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode help.peerColors#f8ed08: field hash: %w", err)
		}
		p.Hash = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode help.peerColors#f8ed08: field colors: %w", err)
		}

		if headerLen > 0 {
			p.Colors = make([]HelpPeerColorOption, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value HelpPeerColorOption
			if err := value.Decode(b); err != nil {
				return fmt.Errorf("unable to decode help.peerColors#f8ed08: field colors: %w", err)
			}
			p.Colors = append(p.Colors, value)
		}
	}
	return nil
}

// GetHash returns value of Hash field.
func (p *HelpPeerColors) GetHash() (value int) {
	if p == nil {
		return
	}
	return p.Hash
}

// GetColors returns value of Colors field.
func (p *HelpPeerColors) GetColors() (value []HelpPeerColorOption) {
	if p == nil {
		return
	}
	return p.Colors
}

// HelpPeerColorsClassName is schema name of HelpPeerColorsClass.
const HelpPeerColorsClassName = "help.PeerColors"

// HelpPeerColorsClass represents help.PeerColors generic type.
//
// See https://core.telegram.org/type/help.PeerColors for reference.
//
// Constructors:
//   - [HelpPeerColorsNotModified]
//   - [HelpPeerColors]
//
// Example:
//
//	g, err := tg.DecodeHelpPeerColors(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tg.HelpPeerColorsNotModified: // help.peerColorsNotModified#2ba1f5ce
//	case *tg.HelpPeerColors: // help.peerColors#f8ed08
//	default: panic(v)
//	}
type HelpPeerColorsClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() HelpPeerColorsClass

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

	// AsModified tries to map HelpPeerColorsClass to HelpPeerColors.
	AsModified() (*HelpPeerColors, bool)
}

// AsModified tries to map HelpPeerColorsNotModified to HelpPeerColors.
func (p *HelpPeerColorsNotModified) AsModified() (*HelpPeerColors, bool) {
	return nil, false
}

// AsModified tries to map HelpPeerColors to HelpPeerColors.
func (p *HelpPeerColors) AsModified() (*HelpPeerColors, bool) {
	return p, true
}

// DecodeHelpPeerColors implements binary de-serialization for HelpPeerColorsClass.
func DecodeHelpPeerColors(buf *bin.Buffer) (HelpPeerColorsClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case HelpPeerColorsNotModifiedTypeID:
		// Decoding help.peerColorsNotModified#2ba1f5ce.
		v := HelpPeerColorsNotModified{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode HelpPeerColorsClass: %w", err)
		}
		return &v, nil
	case HelpPeerColorsTypeID:
		// Decoding help.peerColors#f8ed08.
		v := HelpPeerColors{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode HelpPeerColorsClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode HelpPeerColorsClass: %w", bin.NewUnexpectedID(id))
	}
}

// HelpPeerColors boxes the HelpPeerColorsClass providing a helper.
type HelpPeerColorsBox struct {
	PeerColors HelpPeerColorsClass
}

// Decode implements bin.Decoder for HelpPeerColorsBox.
func (b *HelpPeerColorsBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode HelpPeerColorsBox to nil")
	}
	v, err := DecodeHelpPeerColors(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PeerColors = v
	return nil
}

// Encode implements bin.Encode for HelpPeerColorsBox.
func (b *HelpPeerColorsBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.PeerColors == nil {
		return fmt.Errorf("unable to encode HelpPeerColorsClass as nil")
	}
	return b.PeerColors.Encode(buf)
}
