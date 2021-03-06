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

// MessagesFavedStickersNotModified represents TL type `messages.favedStickersNotModified#9e8fa6d3`.
// No new favorited stickers were found
//
// See https://core.telegram.org/constructor/messages.favedStickersNotModified for reference.
type MessagesFavedStickersNotModified struct {
}

// MessagesFavedStickersNotModifiedTypeID is TL type id of MessagesFavedStickersNotModified.
const MessagesFavedStickersNotModifiedTypeID = 0x9e8fa6d3

func (f *MessagesFavedStickersNotModified) Zero() bool {
	if f == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (f *MessagesFavedStickersNotModified) String() string {
	if f == nil {
		return "MessagesFavedStickersNotModified(nil)"
	}
	type Alias MessagesFavedStickersNotModified
	return fmt.Sprintf("MessagesFavedStickersNotModified%+v", Alias(*f))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesFavedStickersNotModified) TypeID() uint32 {
	return MessagesFavedStickersNotModifiedTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesFavedStickersNotModified) TypeName() string {
	return "messages.favedStickersNotModified"
}

// TypeInfo returns info about TL type.
func (f *MessagesFavedStickersNotModified) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.favedStickersNotModified",
		ID:   MessagesFavedStickersNotModifiedTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (f *MessagesFavedStickersNotModified) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode messages.favedStickersNotModified#9e8fa6d3 as nil")
	}
	b.PutID(MessagesFavedStickersNotModifiedTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *MessagesFavedStickersNotModified) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode messages.favedStickersNotModified#9e8fa6d3 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (f *MessagesFavedStickersNotModified) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode messages.favedStickersNotModified#9e8fa6d3 to nil")
	}
	if err := b.ConsumeID(MessagesFavedStickersNotModifiedTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.favedStickersNotModified#9e8fa6d3: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *MessagesFavedStickersNotModified) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode messages.favedStickersNotModified#9e8fa6d3 to nil")
	}
	return nil
}

// construct implements constructor of MessagesFavedStickersClass.
func (f MessagesFavedStickersNotModified) construct() MessagesFavedStickersClass { return &f }

// Ensuring interfaces in compile-time for MessagesFavedStickersNotModified.
var (
	_ bin.Encoder     = &MessagesFavedStickersNotModified{}
	_ bin.Decoder     = &MessagesFavedStickersNotModified{}
	_ bin.BareEncoder = &MessagesFavedStickersNotModified{}
	_ bin.BareDecoder = &MessagesFavedStickersNotModified{}

	_ MessagesFavedStickersClass = &MessagesFavedStickersNotModified{}
)

// MessagesFavedStickers represents TL type `messages.favedStickers#f37f2f16`.
// Favorited stickers
//
// See https://core.telegram.org/constructor/messages.favedStickers for reference.
type MessagesFavedStickers struct {
	// Hash for pagination, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets#hash-generation
	Hash int
	// Emojis associated to stickers
	Packs []StickerPack
	// Favorited stickers
	Stickers []DocumentClass
}

// MessagesFavedStickersTypeID is TL type id of MessagesFavedStickers.
const MessagesFavedStickersTypeID = 0xf37f2f16

func (f *MessagesFavedStickers) Zero() bool {
	if f == nil {
		return true
	}
	if !(f.Hash == 0) {
		return false
	}
	if !(f.Packs == nil) {
		return false
	}
	if !(f.Stickers == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (f *MessagesFavedStickers) String() string {
	if f == nil {
		return "MessagesFavedStickers(nil)"
	}
	type Alias MessagesFavedStickers
	return fmt.Sprintf("MessagesFavedStickers%+v", Alias(*f))
}

// FillFrom fills MessagesFavedStickers from given interface.
func (f *MessagesFavedStickers) FillFrom(from interface {
	GetHash() (value int)
	GetPacks() (value []StickerPack)
	GetStickers() (value []DocumentClass)
}) {
	f.Hash = from.GetHash()
	f.Packs = from.GetPacks()
	f.Stickers = from.GetStickers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesFavedStickers) TypeID() uint32 {
	return MessagesFavedStickersTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesFavedStickers) TypeName() string {
	return "messages.favedStickers"
}

// TypeInfo returns info about TL type.
func (f *MessagesFavedStickers) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.favedStickers",
		ID:   MessagesFavedStickersTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Hash",
			SchemaName: "hash",
		},
		{
			Name:       "Packs",
			SchemaName: "packs",
		},
		{
			Name:       "Stickers",
			SchemaName: "stickers",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (f *MessagesFavedStickers) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode messages.favedStickers#f37f2f16 as nil")
	}
	b.PutID(MessagesFavedStickersTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *MessagesFavedStickers) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode messages.favedStickers#f37f2f16 as nil")
	}
	b.PutInt(f.Hash)
	b.PutVectorHeader(len(f.Packs))
	for idx, v := range f.Packs {
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode messages.favedStickers#f37f2f16: field packs element with index %d: %w", idx, err)
		}
	}
	b.PutVectorHeader(len(f.Stickers))
	for idx, v := range f.Stickers {
		if v == nil {
			return fmt.Errorf("unable to encode messages.favedStickers#f37f2f16: field stickers element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode messages.favedStickers#f37f2f16: field stickers element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetHash returns value of Hash field.
func (f *MessagesFavedStickers) GetHash() (value int) {
	return f.Hash
}

// GetPacks returns value of Packs field.
func (f *MessagesFavedStickers) GetPacks() (value []StickerPack) {
	return f.Packs
}

// GetStickers returns value of Stickers field.
func (f *MessagesFavedStickers) GetStickers() (value []DocumentClass) {
	return f.Stickers
}

// MapStickers returns field Stickers wrapped in DocumentClassArray helper.
func (f *MessagesFavedStickers) MapStickers() (value DocumentClassArray) {
	return DocumentClassArray(f.Stickers)
}

// Decode implements bin.Decoder.
func (f *MessagesFavedStickers) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode messages.favedStickers#f37f2f16 to nil")
	}
	if err := b.ConsumeID(MessagesFavedStickersTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *MessagesFavedStickers) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode messages.favedStickers#f37f2f16 to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: field hash: %w", err)
		}
		f.Hash = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: field packs: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value StickerPack
			if err := value.Decode(b); err != nil {
				return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: field packs: %w", err)
			}
			f.Packs = append(f.Packs, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: field stickers: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeDocument(b)
			if err != nil {
				return fmt.Errorf("unable to decode messages.favedStickers#f37f2f16: field stickers: %w", err)
			}
			f.Stickers = append(f.Stickers, value)
		}
	}
	return nil
}

// construct implements constructor of MessagesFavedStickersClass.
func (f MessagesFavedStickers) construct() MessagesFavedStickersClass { return &f }

// Ensuring interfaces in compile-time for MessagesFavedStickers.
var (
	_ bin.Encoder     = &MessagesFavedStickers{}
	_ bin.Decoder     = &MessagesFavedStickers{}
	_ bin.BareEncoder = &MessagesFavedStickers{}
	_ bin.BareDecoder = &MessagesFavedStickers{}

	_ MessagesFavedStickersClass = &MessagesFavedStickers{}
)

// MessagesFavedStickersClass represents messages.FavedStickers generic type.
//
// See https://core.telegram.org/type/messages.FavedStickers for reference.
//
// Example:
//  g, err := tg.DecodeMessagesFavedStickers(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.MessagesFavedStickersNotModified: // messages.favedStickersNotModified#9e8fa6d3
//  case *tg.MessagesFavedStickers: // messages.favedStickers#f37f2f16
//  default: panic(v)
//  }
type MessagesFavedStickersClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() MessagesFavedStickersClass

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

	// AsModified tries to map MessagesFavedStickersClass to MessagesFavedStickers.
	AsModified() (*MessagesFavedStickers, bool)
}

// AsModified tries to map MessagesFavedStickersNotModified to MessagesFavedStickers.
func (f *MessagesFavedStickersNotModified) AsModified() (*MessagesFavedStickers, bool) {
	return nil, false
}

// AsModified tries to map MessagesFavedStickers to MessagesFavedStickers.
func (f *MessagesFavedStickers) AsModified() (*MessagesFavedStickers, bool) {
	return f, true
}

// DecodeMessagesFavedStickers implements binary de-serialization for MessagesFavedStickersClass.
func DecodeMessagesFavedStickers(buf *bin.Buffer) (MessagesFavedStickersClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case MessagesFavedStickersNotModifiedTypeID:
		// Decoding messages.favedStickersNotModified#9e8fa6d3.
		v := MessagesFavedStickersNotModified{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode MessagesFavedStickersClass: %w", err)
		}
		return &v, nil
	case MessagesFavedStickersTypeID:
		// Decoding messages.favedStickers#f37f2f16.
		v := MessagesFavedStickers{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode MessagesFavedStickersClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode MessagesFavedStickersClass: %w", bin.NewUnexpectedID(id))
	}
}

// MessagesFavedStickers boxes the MessagesFavedStickersClass providing a helper.
type MessagesFavedStickersBox struct {
	FavedStickers MessagesFavedStickersClass
}

// Decode implements bin.Decoder for MessagesFavedStickersBox.
func (b *MessagesFavedStickersBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode MessagesFavedStickersBox to nil")
	}
	v, err := DecodeMessagesFavedStickers(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.FavedStickers = v
	return nil
}

// Encode implements bin.Encode for MessagesFavedStickersBox.
func (b *MessagesFavedStickersBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.FavedStickers == nil {
		return fmt.Errorf("unable to encode MessagesFavedStickersClass as nil")
	}
	return b.FavedStickers.Encode(buf)
}

// MessagesFavedStickersClassArray is adapter for slice of MessagesFavedStickersClass.
type MessagesFavedStickersClassArray []MessagesFavedStickersClass

// Sort sorts slice of MessagesFavedStickersClass.
func (s MessagesFavedStickersClassArray) Sort(less func(a, b MessagesFavedStickersClass) bool) MessagesFavedStickersClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of MessagesFavedStickersClass.
func (s MessagesFavedStickersClassArray) SortStable(less func(a, b MessagesFavedStickersClass) bool) MessagesFavedStickersClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of MessagesFavedStickersClass.
func (s MessagesFavedStickersClassArray) Retain(keep func(x MessagesFavedStickersClass) bool) MessagesFavedStickersClassArray {
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
func (s MessagesFavedStickersClassArray) First() (v MessagesFavedStickersClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s MessagesFavedStickersClassArray) Last() (v MessagesFavedStickersClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *MessagesFavedStickersClassArray) PopFirst() (v MessagesFavedStickersClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero MessagesFavedStickersClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *MessagesFavedStickersClassArray) Pop() (v MessagesFavedStickersClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsMessagesFavedStickers returns copy with only MessagesFavedStickers constructors.
func (s MessagesFavedStickersClassArray) AsMessagesFavedStickers() (to MessagesFavedStickersArray) {
	for _, elem := range s {
		value, ok := elem.(*MessagesFavedStickers)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AppendOnlyModified appends only Modified constructors to
// given slice.
func (s MessagesFavedStickersClassArray) AppendOnlyModified(to []*MessagesFavedStickers) []*MessagesFavedStickers {
	for _, elem := range s {
		value, ok := elem.AsModified()
		if !ok {
			continue
		}
		to = append(to, value)
	}

	return to
}

// AsModified returns copy with only Modified constructors.
func (s MessagesFavedStickersClassArray) AsModified() (to []*MessagesFavedStickers) {
	return s.AppendOnlyModified(to)
}

// FirstAsModified returns first element of slice (if exists).
func (s MessagesFavedStickersClassArray) FirstAsModified() (v *MessagesFavedStickers, ok bool) {
	value, ok := s.First()
	if !ok {
		return
	}
	return value.AsModified()
}

// LastAsModified returns last element of slice (if exists).
func (s MessagesFavedStickersClassArray) LastAsModified() (v *MessagesFavedStickers, ok bool) {
	value, ok := s.Last()
	if !ok {
		return
	}
	return value.AsModified()
}

// PopFirstAsModified returns element of slice (if exists).
func (s *MessagesFavedStickersClassArray) PopFirstAsModified() (v *MessagesFavedStickers, ok bool) {
	value, ok := s.PopFirst()
	if !ok {
		return
	}
	return value.AsModified()
}

// PopAsModified returns element of slice (if exists).
func (s *MessagesFavedStickersClassArray) PopAsModified() (v *MessagesFavedStickers, ok bool) {
	value, ok := s.Pop()
	if !ok {
		return
	}
	return value.AsModified()
}

// MessagesFavedStickersArray is adapter for slice of MessagesFavedStickers.
type MessagesFavedStickersArray []MessagesFavedStickers

// Sort sorts slice of MessagesFavedStickers.
func (s MessagesFavedStickersArray) Sort(less func(a, b MessagesFavedStickers) bool) MessagesFavedStickersArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of MessagesFavedStickers.
func (s MessagesFavedStickersArray) SortStable(less func(a, b MessagesFavedStickers) bool) MessagesFavedStickersArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of MessagesFavedStickers.
func (s MessagesFavedStickersArray) Retain(keep func(x MessagesFavedStickers) bool) MessagesFavedStickersArray {
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
func (s MessagesFavedStickersArray) First() (v MessagesFavedStickers, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s MessagesFavedStickersArray) Last() (v MessagesFavedStickers, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *MessagesFavedStickersArray) PopFirst() (v MessagesFavedStickers, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero MessagesFavedStickers
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *MessagesFavedStickersArray) Pop() (v MessagesFavedStickers, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
