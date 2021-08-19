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

// PageListItemText represents TL type `pageListItemText#b92fb6cd`.
// List item
//
// See https://core.telegram.org/constructor/pageListItemText for reference.
type PageListItemText struct {
	// Text
	Text RichTextClass
}

// PageListItemTextTypeID is TL type id of PageListItemText.
const PageListItemTextTypeID = 0xb92fb6cd

func (p *PageListItemText) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.Text == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PageListItemText) String() string {
	if p == nil {
		return "PageListItemText(nil)"
	}
	type Alias PageListItemText
	return fmt.Sprintf("PageListItemText%+v", Alias(*p))
}

// FillFrom fills PageListItemText from given interface.
func (p *PageListItemText) FillFrom(from interface {
	GetText() (value RichTextClass)
}) {
	p.Text = from.GetText()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PageListItemText) TypeID() uint32 {
	return PageListItemTextTypeID
}

// TypeName returns name of type in TL schema.
func (*PageListItemText) TypeName() string {
	return "pageListItemText"
}

// TypeInfo returns info about TL type.
func (p *PageListItemText) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "pageListItemText",
		ID:   PageListItemTextTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Text",
			SchemaName: "text",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PageListItemText) Encode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "pageListItemText#b92fb6cd",
		}
	}
	b.PutID(PageListItemTextTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PageListItemText) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "pageListItemText#b92fb6cd",
		}
	}
	if p.Text == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "pageListItemText#b92fb6cd",
			FieldName: "text",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "RichText",
			},
		}
	}
	if err := p.Text.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "pageListItemText#b92fb6cd",
			FieldName:  "text",
			Underlying: err,
		}
	}
	return nil
}

// GetText returns value of Text field.
func (p *PageListItemText) GetText() (value RichTextClass) {
	return p.Text
}

// Decode implements bin.Decoder.
func (p *PageListItemText) Decode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "pageListItemText#b92fb6cd",
		}
	}
	if err := b.ConsumeID(PageListItemTextTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "pageListItemText#b92fb6cd",
			Underlying: err,
		}
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PageListItemText) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "pageListItemText#b92fb6cd",
		}
	}
	{
		value, err := DecodeRichText(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "pageListItemText#b92fb6cd",
				FieldName:  "text",
				Underlying: err,
			}
		}
		p.Text = value
	}
	return nil
}

// construct implements constructor of PageListItemClass.
func (p PageListItemText) construct() PageListItemClass { return &p }

// Ensuring interfaces in compile-time for PageListItemText.
var (
	_ bin.Encoder     = &PageListItemText{}
	_ bin.Decoder     = &PageListItemText{}
	_ bin.BareEncoder = &PageListItemText{}
	_ bin.BareDecoder = &PageListItemText{}

	_ PageListItemClass = &PageListItemText{}
)

// PageListItemBlocks represents TL type `pageListItemBlocks#25e073fc`.
// List item
//
// See https://core.telegram.org/constructor/pageListItemBlocks for reference.
type PageListItemBlocks struct {
	// Blocks
	Blocks []PageBlockClass
}

// PageListItemBlocksTypeID is TL type id of PageListItemBlocks.
const PageListItemBlocksTypeID = 0x25e073fc

func (p *PageListItemBlocks) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.Blocks == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PageListItemBlocks) String() string {
	if p == nil {
		return "PageListItemBlocks(nil)"
	}
	type Alias PageListItemBlocks
	return fmt.Sprintf("PageListItemBlocks%+v", Alias(*p))
}

// FillFrom fills PageListItemBlocks from given interface.
func (p *PageListItemBlocks) FillFrom(from interface {
	GetBlocks() (value []PageBlockClass)
}) {
	p.Blocks = from.GetBlocks()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PageListItemBlocks) TypeID() uint32 {
	return PageListItemBlocksTypeID
}

// TypeName returns name of type in TL schema.
func (*PageListItemBlocks) TypeName() string {
	return "pageListItemBlocks"
}

// TypeInfo returns info about TL type.
func (p *PageListItemBlocks) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "pageListItemBlocks",
		ID:   PageListItemBlocksTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Blocks",
			SchemaName: "blocks",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PageListItemBlocks) Encode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "pageListItemBlocks#25e073fc",
		}
	}
	b.PutID(PageListItemBlocksTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PageListItemBlocks) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "pageListItemBlocks#25e073fc",
		}
	}
	b.PutVectorHeader(len(p.Blocks))
	for idx, v := range p.Blocks {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "pageListItemBlocks#25e073fc",
				FieldName: "blocks",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<PageBlock>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "pageListItemBlocks#25e073fc",
				FieldName: "blocks",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	return nil
}

// GetBlocks returns value of Blocks field.
func (p *PageListItemBlocks) GetBlocks() (value []PageBlockClass) {
	return p.Blocks
}

// MapBlocks returns field Blocks wrapped in PageBlockClassArray helper.
func (p *PageListItemBlocks) MapBlocks() (value PageBlockClassArray) {
	return PageBlockClassArray(p.Blocks)
}

// Decode implements bin.Decoder.
func (p *PageListItemBlocks) Decode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "pageListItemBlocks#25e073fc",
		}
	}
	if err := b.ConsumeID(PageListItemBlocksTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "pageListItemBlocks#25e073fc",
			Underlying: err,
		}
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PageListItemBlocks) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "pageListItemBlocks#25e073fc",
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "pageListItemBlocks#25e073fc",
				FieldName:  "blocks",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			p.Blocks = make([]PageBlockClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodePageBlock(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "pageListItemBlocks#25e073fc",
					FieldName:  "blocks",
					Underlying: err,
				}
			}
			p.Blocks = append(p.Blocks, value)
		}
	}
	return nil
}

// construct implements constructor of PageListItemClass.
func (p PageListItemBlocks) construct() PageListItemClass { return &p }

// Ensuring interfaces in compile-time for PageListItemBlocks.
var (
	_ bin.Encoder     = &PageListItemBlocks{}
	_ bin.Decoder     = &PageListItemBlocks{}
	_ bin.BareEncoder = &PageListItemBlocks{}
	_ bin.BareDecoder = &PageListItemBlocks{}

	_ PageListItemClass = &PageListItemBlocks{}
)

// PageListItemClass represents PageListItem generic type.
//
// See https://core.telegram.org/type/PageListItem for reference.
//
// Example:
//  g, err := tg.DecodePageListItem(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.PageListItemText: // pageListItemText#b92fb6cd
//  case *tg.PageListItemBlocks: // pageListItemBlocks#25e073fc
//  default: panic(v)
//  }
type PageListItemClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() PageListItemClass

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

// DecodePageListItem implements binary de-serialization for PageListItemClass.
func DecodePageListItem(buf *bin.Buffer) (PageListItemClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case PageListItemTextTypeID:
		// Decoding pageListItemText#b92fb6cd.
		v := PageListItemText{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "PageListItemClass",
				Underlying: err,
			}
		}
		return &v, nil
	case PageListItemBlocksTypeID:
		// Decoding pageListItemBlocks#25e073fc.
		v := PageListItemBlocks{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "PageListItemClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "PageListItemClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// PageListItem boxes the PageListItemClass providing a helper.
type PageListItemBox struct {
	PageListItem PageListItemClass
}

// Decode implements bin.Decoder for PageListItemBox.
func (b *PageListItemBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "PageListItemBox",
		}
	}
	v, err := DecodePageListItem(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PageListItem = v
	return nil
}

// Encode implements bin.Encode for PageListItemBox.
func (b *PageListItemBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.PageListItem == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "PageListItemBox",
		}
	}
	return b.PageListItem.Encode(buf)
}

// PageListItemClassArray is adapter for slice of PageListItemClass.
type PageListItemClassArray []PageListItemClass

// Sort sorts slice of PageListItemClass.
func (s PageListItemClassArray) Sort(less func(a, b PageListItemClass) bool) PageListItemClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of PageListItemClass.
func (s PageListItemClassArray) SortStable(less func(a, b PageListItemClass) bool) PageListItemClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of PageListItemClass.
func (s PageListItemClassArray) Retain(keep func(x PageListItemClass) bool) PageListItemClassArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s PageListItemClassArray) First() (v PageListItemClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s PageListItemClassArray) Last() (v PageListItemClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *PageListItemClassArray) PopFirst() (v PageListItemClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero PageListItemClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *PageListItemClassArray) Pop() (v PageListItemClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsPageListItemText returns copy with only PageListItemText constructors.
func (s PageListItemClassArray) AsPageListItemText() (to PageListItemTextArray) {
	for _, elem := range s {
		value, ok := elem.(*PageListItemText)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsPageListItemBlocks returns copy with only PageListItemBlocks constructors.
func (s PageListItemClassArray) AsPageListItemBlocks() (to PageListItemBlocksArray) {
	for _, elem := range s {
		value, ok := elem.(*PageListItemBlocks)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// PageListItemTextArray is adapter for slice of PageListItemText.
type PageListItemTextArray []PageListItemText

// Sort sorts slice of PageListItemText.
func (s PageListItemTextArray) Sort(less func(a, b PageListItemText) bool) PageListItemTextArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of PageListItemText.
func (s PageListItemTextArray) SortStable(less func(a, b PageListItemText) bool) PageListItemTextArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of PageListItemText.
func (s PageListItemTextArray) Retain(keep func(x PageListItemText) bool) PageListItemTextArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s PageListItemTextArray) First() (v PageListItemText, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s PageListItemTextArray) Last() (v PageListItemText, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *PageListItemTextArray) PopFirst() (v PageListItemText, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero PageListItemText
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *PageListItemTextArray) Pop() (v PageListItemText, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// PageListItemBlocksArray is adapter for slice of PageListItemBlocks.
type PageListItemBlocksArray []PageListItemBlocks

// Sort sorts slice of PageListItemBlocks.
func (s PageListItemBlocksArray) Sort(less func(a, b PageListItemBlocks) bool) PageListItemBlocksArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of PageListItemBlocks.
func (s PageListItemBlocksArray) SortStable(less func(a, b PageListItemBlocks) bool) PageListItemBlocksArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of PageListItemBlocks.
func (s PageListItemBlocksArray) Retain(keep func(x PageListItemBlocks) bool) PageListItemBlocksArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s PageListItemBlocksArray) First() (v PageListItemBlocks, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s PageListItemBlocksArray) Last() (v PageListItemBlocks, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *PageListItemBlocksArray) PopFirst() (v PageListItemBlocks, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero PageListItemBlocks
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *PageListItemBlocksArray) Pop() (v PageListItemBlocks, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
