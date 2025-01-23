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

// EmojiStatusEmpty represents TL type `emojiStatusEmpty#2de11aae`.
// No emoji status is set
//
// See https://core.telegram.org/constructor/emojiStatusEmpty for reference.
type EmojiStatusEmpty struct {
}

// EmojiStatusEmptyTypeID is TL type id of EmojiStatusEmpty.
const EmojiStatusEmptyTypeID = 0x2de11aae

// construct implements constructor of EmojiStatusClass.
func (e EmojiStatusEmpty) construct() EmojiStatusClass { return &e }

// Ensuring interfaces in compile-time for EmojiStatusEmpty.
var (
	_ bin.Encoder     = &EmojiStatusEmpty{}
	_ bin.Decoder     = &EmojiStatusEmpty{}
	_ bin.BareEncoder = &EmojiStatusEmpty{}
	_ bin.BareDecoder = &EmojiStatusEmpty{}

	_ EmojiStatusClass = &EmojiStatusEmpty{}
)

func (e *EmojiStatusEmpty) Zero() bool {
	if e == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (e *EmojiStatusEmpty) String() string {
	if e == nil {
		return "EmojiStatusEmpty(nil)"
	}
	type Alias EmojiStatusEmpty
	return fmt.Sprintf("EmojiStatusEmpty%+v", Alias(*e))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*EmojiStatusEmpty) TypeID() uint32 {
	return EmojiStatusEmptyTypeID
}

// TypeName returns name of type in TL schema.
func (*EmojiStatusEmpty) TypeName() string {
	return "emojiStatusEmpty"
}

// TypeInfo returns info about TL type.
func (e *EmojiStatusEmpty) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "emojiStatusEmpty",
		ID:   EmojiStatusEmptyTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (e *EmojiStatusEmpty) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatusEmpty#2de11aae as nil")
	}
	b.PutID(EmojiStatusEmptyTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *EmojiStatusEmpty) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatusEmpty#2de11aae as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *EmojiStatusEmpty) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatusEmpty#2de11aae to nil")
	}
	if err := b.ConsumeID(EmojiStatusEmptyTypeID); err != nil {
		return fmt.Errorf("unable to decode emojiStatusEmpty#2de11aae: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *EmojiStatusEmpty) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatusEmpty#2de11aae to nil")
	}
	return nil
}

// EmojiStatus represents TL type `emojiStatus#e7ff068a`.
// An emoji status¹
//
// Links:
//  1. https://core.telegram.org/api/emoji-status
//
// See https://core.telegram.org/constructor/emojiStatus for reference.
type EmojiStatus struct {
	// Flags field of EmojiStatus.
	Flags bin.Fields
	// Custom emoji document ID¹
	//
	// Links:
	//  1) https://core.telegram.org/api/custom-emoji
	DocumentID int64
	// Until field of EmojiStatus.
	//
	// Use SetUntil and GetUntil helpers.
	Until int
}

// EmojiStatusTypeID is TL type id of EmojiStatus.
const EmojiStatusTypeID = 0xe7ff068a

// construct implements constructor of EmojiStatusClass.
func (e EmojiStatus) construct() EmojiStatusClass { return &e }

// Ensuring interfaces in compile-time for EmojiStatus.
var (
	_ bin.Encoder     = &EmojiStatus{}
	_ bin.Decoder     = &EmojiStatus{}
	_ bin.BareEncoder = &EmojiStatus{}
	_ bin.BareDecoder = &EmojiStatus{}

	_ EmojiStatusClass = &EmojiStatus{}
)

func (e *EmojiStatus) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Flags.Zero()) {
		return false
	}
	if !(e.DocumentID == 0) {
		return false
	}
	if !(e.Until == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *EmojiStatus) String() string {
	if e == nil {
		return "EmojiStatus(nil)"
	}
	type Alias EmojiStatus
	return fmt.Sprintf("EmojiStatus%+v", Alias(*e))
}

// FillFrom fills EmojiStatus from given interface.
func (e *EmojiStatus) FillFrom(from interface {
	GetDocumentID() (value int64)
	GetUntil() (value int, ok bool)
}) {
	e.DocumentID = from.GetDocumentID()
	if val, ok := from.GetUntil(); ok {
		e.Until = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*EmojiStatus) TypeID() uint32 {
	return EmojiStatusTypeID
}

// TypeName returns name of type in TL schema.
func (*EmojiStatus) TypeName() string {
	return "emojiStatus"
}

// TypeInfo returns info about TL type.
func (e *EmojiStatus) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "emojiStatus",
		ID:   EmojiStatusTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "DocumentID",
			SchemaName: "document_id",
		},
		{
			Name:       "Until",
			SchemaName: "until",
			Null:       !e.Flags.Has(0),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (e *EmojiStatus) SetFlags() {
	if !(e.Until == 0) {
		e.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (e *EmojiStatus) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatus#e7ff068a as nil")
	}
	b.PutID(EmojiStatusTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *EmojiStatus) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatus#e7ff068a as nil")
	}
	e.SetFlags()
	if err := e.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode emojiStatus#e7ff068a: field flags: %w", err)
	}
	b.PutLong(e.DocumentID)
	if e.Flags.Has(0) {
		b.PutInt(e.Until)
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *EmojiStatus) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatus#e7ff068a to nil")
	}
	if err := b.ConsumeID(EmojiStatusTypeID); err != nil {
		return fmt.Errorf("unable to decode emojiStatus#e7ff068a: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *EmojiStatus) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatus#e7ff068a to nil")
	}
	{
		if err := e.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode emojiStatus#e7ff068a: field flags: %w", err)
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatus#e7ff068a: field document_id: %w", err)
		}
		e.DocumentID = value
	}
	if e.Flags.Has(0) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatus#e7ff068a: field until: %w", err)
		}
		e.Until = value
	}
	return nil
}

// GetDocumentID returns value of DocumentID field.
func (e *EmojiStatus) GetDocumentID() (value int64) {
	if e == nil {
		return
	}
	return e.DocumentID
}

// SetUntil sets value of Until conditional field.
func (e *EmojiStatus) SetUntil(value int) {
	e.Flags.Set(0)
	e.Until = value
}

// GetUntil returns value of Until conditional field and
// boolean which is true if field was set.
func (e *EmojiStatus) GetUntil() (value int, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(0) {
		return value, false
	}
	return e.Until, true
}

// EmojiStatusCollectible represents TL type `emojiStatusCollectible#7184603b`.
//
// See https://core.telegram.org/constructor/emojiStatusCollectible for reference.
type EmojiStatusCollectible struct {
	// Flags field of EmojiStatusCollectible.
	Flags bin.Fields
	// CollectibleID field of EmojiStatusCollectible.
	CollectibleID int64
	// DocumentID field of EmojiStatusCollectible.
	DocumentID int64
	// Title field of EmojiStatusCollectible.
	Title string
	// Slug field of EmojiStatusCollectible.
	Slug string
	// PatternDocumentID field of EmojiStatusCollectible.
	PatternDocumentID int64
	// CenterColor field of EmojiStatusCollectible.
	CenterColor int
	// EdgeColor field of EmojiStatusCollectible.
	EdgeColor int
	// PatternColor field of EmojiStatusCollectible.
	PatternColor int
	// TextColor field of EmojiStatusCollectible.
	TextColor int
	// Until field of EmojiStatusCollectible.
	//
	// Use SetUntil and GetUntil helpers.
	Until int
}

// EmojiStatusCollectibleTypeID is TL type id of EmojiStatusCollectible.
const EmojiStatusCollectibleTypeID = 0x7184603b

// construct implements constructor of EmojiStatusClass.
func (e EmojiStatusCollectible) construct() EmojiStatusClass { return &e }

// Ensuring interfaces in compile-time for EmojiStatusCollectible.
var (
	_ bin.Encoder     = &EmojiStatusCollectible{}
	_ bin.Decoder     = &EmojiStatusCollectible{}
	_ bin.BareEncoder = &EmojiStatusCollectible{}
	_ bin.BareDecoder = &EmojiStatusCollectible{}

	_ EmojiStatusClass = &EmojiStatusCollectible{}
)

func (e *EmojiStatusCollectible) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Flags.Zero()) {
		return false
	}
	if !(e.CollectibleID == 0) {
		return false
	}
	if !(e.DocumentID == 0) {
		return false
	}
	if !(e.Title == "") {
		return false
	}
	if !(e.Slug == "") {
		return false
	}
	if !(e.PatternDocumentID == 0) {
		return false
	}
	if !(e.CenterColor == 0) {
		return false
	}
	if !(e.EdgeColor == 0) {
		return false
	}
	if !(e.PatternColor == 0) {
		return false
	}
	if !(e.TextColor == 0) {
		return false
	}
	if !(e.Until == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *EmojiStatusCollectible) String() string {
	if e == nil {
		return "EmojiStatusCollectible(nil)"
	}
	type Alias EmojiStatusCollectible
	return fmt.Sprintf("EmojiStatusCollectible%+v", Alias(*e))
}

// FillFrom fills EmojiStatusCollectible from given interface.
func (e *EmojiStatusCollectible) FillFrom(from interface {
	GetCollectibleID() (value int64)
	GetDocumentID() (value int64)
	GetTitle() (value string)
	GetSlug() (value string)
	GetPatternDocumentID() (value int64)
	GetCenterColor() (value int)
	GetEdgeColor() (value int)
	GetPatternColor() (value int)
	GetTextColor() (value int)
	GetUntil() (value int, ok bool)
}) {
	e.CollectibleID = from.GetCollectibleID()
	e.DocumentID = from.GetDocumentID()
	e.Title = from.GetTitle()
	e.Slug = from.GetSlug()
	e.PatternDocumentID = from.GetPatternDocumentID()
	e.CenterColor = from.GetCenterColor()
	e.EdgeColor = from.GetEdgeColor()
	e.PatternColor = from.GetPatternColor()
	e.TextColor = from.GetTextColor()
	if val, ok := from.GetUntil(); ok {
		e.Until = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*EmojiStatusCollectible) TypeID() uint32 {
	return EmojiStatusCollectibleTypeID
}

// TypeName returns name of type in TL schema.
func (*EmojiStatusCollectible) TypeName() string {
	return "emojiStatusCollectible"
}

// TypeInfo returns info about TL type.
func (e *EmojiStatusCollectible) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "emojiStatusCollectible",
		ID:   EmojiStatusCollectibleTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "CollectibleID",
			SchemaName: "collectible_id",
		},
		{
			Name:       "DocumentID",
			SchemaName: "document_id",
		},
		{
			Name:       "Title",
			SchemaName: "title",
		},
		{
			Name:       "Slug",
			SchemaName: "slug",
		},
		{
			Name:       "PatternDocumentID",
			SchemaName: "pattern_document_id",
		},
		{
			Name:       "CenterColor",
			SchemaName: "center_color",
		},
		{
			Name:       "EdgeColor",
			SchemaName: "edge_color",
		},
		{
			Name:       "PatternColor",
			SchemaName: "pattern_color",
		},
		{
			Name:       "TextColor",
			SchemaName: "text_color",
		},
		{
			Name:       "Until",
			SchemaName: "until",
			Null:       !e.Flags.Has(0),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (e *EmojiStatusCollectible) SetFlags() {
	if !(e.Until == 0) {
		e.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (e *EmojiStatusCollectible) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatusCollectible#7184603b as nil")
	}
	b.PutID(EmojiStatusCollectibleTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *EmojiStatusCollectible) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode emojiStatusCollectible#7184603b as nil")
	}
	e.SetFlags()
	if err := e.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode emojiStatusCollectible#7184603b: field flags: %w", err)
	}
	b.PutLong(e.CollectibleID)
	b.PutLong(e.DocumentID)
	b.PutString(e.Title)
	b.PutString(e.Slug)
	b.PutLong(e.PatternDocumentID)
	b.PutInt(e.CenterColor)
	b.PutInt(e.EdgeColor)
	b.PutInt(e.PatternColor)
	b.PutInt(e.TextColor)
	if e.Flags.Has(0) {
		b.PutInt(e.Until)
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *EmojiStatusCollectible) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatusCollectible#7184603b to nil")
	}
	if err := b.ConsumeID(EmojiStatusCollectibleTypeID); err != nil {
		return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *EmojiStatusCollectible) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode emojiStatusCollectible#7184603b to nil")
	}
	{
		if err := e.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field flags: %w", err)
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field collectible_id: %w", err)
		}
		e.CollectibleID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field document_id: %w", err)
		}
		e.DocumentID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field title: %w", err)
		}
		e.Title = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field slug: %w", err)
		}
		e.Slug = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field pattern_document_id: %w", err)
		}
		e.PatternDocumentID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field center_color: %w", err)
		}
		e.CenterColor = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field edge_color: %w", err)
		}
		e.EdgeColor = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field pattern_color: %w", err)
		}
		e.PatternColor = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field text_color: %w", err)
		}
		e.TextColor = value
	}
	if e.Flags.Has(0) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode emojiStatusCollectible#7184603b: field until: %w", err)
		}
		e.Until = value
	}
	return nil
}

// GetCollectibleID returns value of CollectibleID field.
func (e *EmojiStatusCollectible) GetCollectibleID() (value int64) {
	if e == nil {
		return
	}
	return e.CollectibleID
}

// GetDocumentID returns value of DocumentID field.
func (e *EmojiStatusCollectible) GetDocumentID() (value int64) {
	if e == nil {
		return
	}
	return e.DocumentID
}

// GetTitle returns value of Title field.
func (e *EmojiStatusCollectible) GetTitle() (value string) {
	if e == nil {
		return
	}
	return e.Title
}

// GetSlug returns value of Slug field.
func (e *EmojiStatusCollectible) GetSlug() (value string) {
	if e == nil {
		return
	}
	return e.Slug
}

// GetPatternDocumentID returns value of PatternDocumentID field.
func (e *EmojiStatusCollectible) GetPatternDocumentID() (value int64) {
	if e == nil {
		return
	}
	return e.PatternDocumentID
}

// GetCenterColor returns value of CenterColor field.
func (e *EmojiStatusCollectible) GetCenterColor() (value int) {
	if e == nil {
		return
	}
	return e.CenterColor
}

// GetEdgeColor returns value of EdgeColor field.
func (e *EmojiStatusCollectible) GetEdgeColor() (value int) {
	if e == nil {
		return
	}
	return e.EdgeColor
}

// GetPatternColor returns value of PatternColor field.
func (e *EmojiStatusCollectible) GetPatternColor() (value int) {
	if e == nil {
		return
	}
	return e.PatternColor
}

// GetTextColor returns value of TextColor field.
func (e *EmojiStatusCollectible) GetTextColor() (value int) {
	if e == nil {
		return
	}
	return e.TextColor
}

// SetUntil sets value of Until conditional field.
func (e *EmojiStatusCollectible) SetUntil(value int) {
	e.Flags.Set(0)
	e.Until = value
}

// GetUntil returns value of Until conditional field and
// boolean which is true if field was set.
func (e *EmojiStatusCollectible) GetUntil() (value int, ok bool) {
	if e == nil {
		return
	}
	if !e.Flags.Has(0) {
		return value, false
	}
	return e.Until, true
}

// InputEmojiStatusCollectible represents TL type `inputEmojiStatusCollectible#7141dbf`.
//
// See https://core.telegram.org/constructor/inputEmojiStatusCollectible for reference.
type InputEmojiStatusCollectible struct {
	// Flags field of InputEmojiStatusCollectible.
	Flags bin.Fields
	// CollectibleID field of InputEmojiStatusCollectible.
	CollectibleID int64
	// Until field of InputEmojiStatusCollectible.
	//
	// Use SetUntil and GetUntil helpers.
	Until int
}

// InputEmojiStatusCollectibleTypeID is TL type id of InputEmojiStatusCollectible.
const InputEmojiStatusCollectibleTypeID = 0x7141dbf

// construct implements constructor of EmojiStatusClass.
func (i InputEmojiStatusCollectible) construct() EmojiStatusClass { return &i }

// Ensuring interfaces in compile-time for InputEmojiStatusCollectible.
var (
	_ bin.Encoder     = &InputEmojiStatusCollectible{}
	_ bin.Decoder     = &InputEmojiStatusCollectible{}
	_ bin.BareEncoder = &InputEmojiStatusCollectible{}
	_ bin.BareDecoder = &InputEmojiStatusCollectible{}

	_ EmojiStatusClass = &InputEmojiStatusCollectible{}
)

func (i *InputEmojiStatusCollectible) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Flags.Zero()) {
		return false
	}
	if !(i.CollectibleID == 0) {
		return false
	}
	if !(i.Until == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputEmojiStatusCollectible) String() string {
	if i == nil {
		return "InputEmojiStatusCollectible(nil)"
	}
	type Alias InputEmojiStatusCollectible
	return fmt.Sprintf("InputEmojiStatusCollectible%+v", Alias(*i))
}

// FillFrom fills InputEmojiStatusCollectible from given interface.
func (i *InputEmojiStatusCollectible) FillFrom(from interface {
	GetCollectibleID() (value int64)
	GetUntil() (value int, ok bool)
}) {
	i.CollectibleID = from.GetCollectibleID()
	if val, ok := from.GetUntil(); ok {
		i.Until = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputEmojiStatusCollectible) TypeID() uint32 {
	return InputEmojiStatusCollectibleTypeID
}

// TypeName returns name of type in TL schema.
func (*InputEmojiStatusCollectible) TypeName() string {
	return "inputEmojiStatusCollectible"
}

// TypeInfo returns info about TL type.
func (i *InputEmojiStatusCollectible) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputEmojiStatusCollectible",
		ID:   InputEmojiStatusCollectibleTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "CollectibleID",
			SchemaName: "collectible_id",
		},
		{
			Name:       "Until",
			SchemaName: "until",
			Null:       !i.Flags.Has(0),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (i *InputEmojiStatusCollectible) SetFlags() {
	if !(i.Until == 0) {
		i.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (i *InputEmojiStatusCollectible) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputEmojiStatusCollectible#7141dbf as nil")
	}
	b.PutID(InputEmojiStatusCollectibleTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputEmojiStatusCollectible) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputEmojiStatusCollectible#7141dbf as nil")
	}
	i.SetFlags()
	if err := i.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputEmojiStatusCollectible#7141dbf: field flags: %w", err)
	}
	b.PutLong(i.CollectibleID)
	if i.Flags.Has(0) {
		b.PutInt(i.Until)
	}
	return nil
}

// Decode implements bin.Decoder.
func (i *InputEmojiStatusCollectible) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputEmojiStatusCollectible#7141dbf to nil")
	}
	if err := b.ConsumeID(InputEmojiStatusCollectibleTypeID); err != nil {
		return fmt.Errorf("unable to decode inputEmojiStatusCollectible#7141dbf: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputEmojiStatusCollectible) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputEmojiStatusCollectible#7141dbf to nil")
	}
	{
		if err := i.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputEmojiStatusCollectible#7141dbf: field flags: %w", err)
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode inputEmojiStatusCollectible#7141dbf: field collectible_id: %w", err)
		}
		i.CollectibleID = value
	}
	if i.Flags.Has(0) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputEmojiStatusCollectible#7141dbf: field until: %w", err)
		}
		i.Until = value
	}
	return nil
}

// GetCollectibleID returns value of CollectibleID field.
func (i *InputEmojiStatusCollectible) GetCollectibleID() (value int64) {
	if i == nil {
		return
	}
	return i.CollectibleID
}

// SetUntil sets value of Until conditional field.
func (i *InputEmojiStatusCollectible) SetUntil(value int) {
	i.Flags.Set(0)
	i.Until = value
}

// GetUntil returns value of Until conditional field and
// boolean which is true if field was set.
func (i *InputEmojiStatusCollectible) GetUntil() (value int, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(0) {
		return value, false
	}
	return i.Until, true
}

// EmojiStatusClassName is schema name of EmojiStatusClass.
const EmojiStatusClassName = "EmojiStatus"

// EmojiStatusClass represents EmojiStatus generic type.
//
// See https://core.telegram.org/type/EmojiStatus for reference.
//
// Example:
//
//	g, err := tg.DecodeEmojiStatus(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tg.EmojiStatusEmpty: // emojiStatusEmpty#2de11aae
//	case *tg.EmojiStatus: // emojiStatus#e7ff068a
//	case *tg.EmojiStatusCollectible: // emojiStatusCollectible#7184603b
//	case *tg.InputEmojiStatusCollectible: // inputEmojiStatusCollectible#7141dbf
//	default: panic(v)
//	}
type EmojiStatusClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() EmojiStatusClass

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

	// AsNotEmpty tries to map EmojiStatusClass to NotEmptyEmojiStatus.
	AsNotEmpty() (NotEmptyEmojiStatus, bool)
}

// NotEmptyEmojiStatus represents NotEmpty subset of EmojiStatusClass.
type NotEmptyEmojiStatus interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() EmojiStatusClass

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

	// Until field of EmojiStatus.
	GetUntil() (value int, ok bool)
}

// AsNotEmpty tries to map EmojiStatusEmpty to NotEmptyEmojiStatus.
func (e *EmojiStatusEmpty) AsNotEmpty() (NotEmptyEmojiStatus, bool) {
	value, ok := (EmojiStatusClass(e)).(NotEmptyEmojiStatus)
	return value, ok
}

// AsNotEmpty tries to map EmojiStatus to NotEmptyEmojiStatus.
func (e *EmojiStatus) AsNotEmpty() (NotEmptyEmojiStatus, bool) {
	value, ok := (EmojiStatusClass(e)).(NotEmptyEmojiStatus)
	return value, ok
}

// AsNotEmpty tries to map EmojiStatusCollectible to NotEmptyEmojiStatus.
func (e *EmojiStatusCollectible) AsNotEmpty() (NotEmptyEmojiStatus, bool) {
	value, ok := (EmojiStatusClass(e)).(NotEmptyEmojiStatus)
	return value, ok
}

// AsNotEmpty tries to map InputEmojiStatusCollectible to NotEmptyEmojiStatus.
func (i *InputEmojiStatusCollectible) AsNotEmpty() (NotEmptyEmojiStatus, bool) {
	value, ok := (EmojiStatusClass(i)).(NotEmptyEmojiStatus)
	return value, ok
}

// DecodeEmojiStatus implements binary de-serialization for EmojiStatusClass.
func DecodeEmojiStatus(buf *bin.Buffer) (EmojiStatusClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case EmojiStatusEmptyTypeID:
		// Decoding emojiStatusEmpty#2de11aae.
		v := EmojiStatusEmpty{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode EmojiStatusClass: %w", err)
		}
		return &v, nil
	case EmojiStatusTypeID:
		// Decoding emojiStatus#e7ff068a.
		v := EmojiStatus{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode EmojiStatusClass: %w", err)
		}
		return &v, nil
	case EmojiStatusCollectibleTypeID:
		// Decoding emojiStatusCollectible#7184603b.
		v := EmojiStatusCollectible{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode EmojiStatusClass: %w", err)
		}
		return &v, nil
	case InputEmojiStatusCollectibleTypeID:
		// Decoding inputEmojiStatusCollectible#7141dbf.
		v := InputEmojiStatusCollectible{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode EmojiStatusClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode EmojiStatusClass: %w", bin.NewUnexpectedID(id))
	}
}

// EmojiStatus boxes the EmojiStatusClass providing a helper.
type EmojiStatusBox struct {
	EmojiStatus EmojiStatusClass
}

// Decode implements bin.Decoder for EmojiStatusBox.
func (b *EmojiStatusBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode EmojiStatusBox to nil")
	}
	v, err := DecodeEmojiStatus(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.EmojiStatus = v
	return nil
}

// Encode implements bin.Encode for EmojiStatusBox.
func (b *EmojiStatusBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.EmojiStatus == nil {
		return fmt.Errorf("unable to encode EmojiStatusClass as nil")
	}
	return b.EmojiStatus.Encode(buf)
}
