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

// InputInvoiceMessage represents TL type `inputInvoiceMessage#c5b56859`.
// An invoice contained in a messageMediaInvoice¹ message.
//
// Links:
//  1. https://core.telegram.org/constructor/messageMediaInvoice
//
// See https://core.telegram.org/constructor/inputInvoiceMessage for reference.
type InputInvoiceMessage struct {
	// Chat where the invoice was sent
	Peer InputPeerClass
	// Message ID
	MsgID int
}

// InputInvoiceMessageTypeID is TL type id of InputInvoiceMessage.
const InputInvoiceMessageTypeID = 0xc5b56859

// construct implements constructor of InputInvoiceClass.
func (i InputInvoiceMessage) construct() InputInvoiceClass { return &i }

// Ensuring interfaces in compile-time for InputInvoiceMessage.
var (
	_ bin.Encoder     = &InputInvoiceMessage{}
	_ bin.Decoder     = &InputInvoiceMessage{}
	_ bin.BareEncoder = &InputInvoiceMessage{}
	_ bin.BareDecoder = &InputInvoiceMessage{}

	_ InputInvoiceClass = &InputInvoiceMessage{}
)

func (i *InputInvoiceMessage) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Peer == nil) {
		return false
	}
	if !(i.MsgID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputInvoiceMessage) String() string {
	if i == nil {
		return "InputInvoiceMessage(nil)"
	}
	type Alias InputInvoiceMessage
	return fmt.Sprintf("InputInvoiceMessage%+v", Alias(*i))
}

// FillFrom fills InputInvoiceMessage from given interface.
func (i *InputInvoiceMessage) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
	GetMsgID() (value int)
}) {
	i.Peer = from.GetPeer()
	i.MsgID = from.GetMsgID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputInvoiceMessage) TypeID() uint32 {
	return InputInvoiceMessageTypeID
}

// TypeName returns name of type in TL schema.
func (*InputInvoiceMessage) TypeName() string {
	return "inputInvoiceMessage"
}

// TypeInfo returns info about TL type.
func (i *InputInvoiceMessage) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputInvoiceMessage",
		ID:   InputInvoiceMessageTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "MsgID",
			SchemaName: "msg_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputInvoiceMessage) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoiceMessage#c5b56859 as nil")
	}
	b.PutID(InputInvoiceMessageTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputInvoiceMessage) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoiceMessage#c5b56859 as nil")
	}
	if i.Peer == nil {
		return fmt.Errorf("unable to encode inputInvoiceMessage#c5b56859: field peer is nil")
	}
	if err := i.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputInvoiceMessage#c5b56859: field peer: %w", err)
	}
	b.PutInt(i.MsgID)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputInvoiceMessage) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoiceMessage#c5b56859 to nil")
	}
	if err := b.ConsumeID(InputInvoiceMessageTypeID); err != nil {
		return fmt.Errorf("unable to decode inputInvoiceMessage#c5b56859: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputInvoiceMessage) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoiceMessage#c5b56859 to nil")
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode inputInvoiceMessage#c5b56859: field peer: %w", err)
		}
		i.Peer = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputInvoiceMessage#c5b56859: field msg_id: %w", err)
		}
		i.MsgID = value
	}
	return nil
}

// GetPeer returns value of Peer field.
func (i *InputInvoiceMessage) GetPeer() (value InputPeerClass) {
	if i == nil {
		return
	}
	return i.Peer
}

// GetMsgID returns value of MsgID field.
func (i *InputInvoiceMessage) GetMsgID() (value int) {
	if i == nil {
		return
	}
	return i.MsgID
}

// InputInvoiceSlug represents TL type `inputInvoiceSlug#c326caef`.
// An invoice slug taken from an invoice deep link¹ or from the premium_invoice_slug app
// config parameter »²
//
// Links:
//  1. https://core.telegram.org/api/links#invoice-links
//  2. https://core.telegram.org/api/config#premium-invoice-slug
//
// See https://core.telegram.org/constructor/inputInvoiceSlug for reference.
type InputInvoiceSlug struct {
	// The invoice slug
	Slug string
}

// InputInvoiceSlugTypeID is TL type id of InputInvoiceSlug.
const InputInvoiceSlugTypeID = 0xc326caef

// construct implements constructor of InputInvoiceClass.
func (i InputInvoiceSlug) construct() InputInvoiceClass { return &i }

// Ensuring interfaces in compile-time for InputInvoiceSlug.
var (
	_ bin.Encoder     = &InputInvoiceSlug{}
	_ bin.Decoder     = &InputInvoiceSlug{}
	_ bin.BareEncoder = &InputInvoiceSlug{}
	_ bin.BareDecoder = &InputInvoiceSlug{}

	_ InputInvoiceClass = &InputInvoiceSlug{}
)

func (i *InputInvoiceSlug) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Slug == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputInvoiceSlug) String() string {
	if i == nil {
		return "InputInvoiceSlug(nil)"
	}
	type Alias InputInvoiceSlug
	return fmt.Sprintf("InputInvoiceSlug%+v", Alias(*i))
}

// FillFrom fills InputInvoiceSlug from given interface.
func (i *InputInvoiceSlug) FillFrom(from interface {
	GetSlug() (value string)
}) {
	i.Slug = from.GetSlug()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputInvoiceSlug) TypeID() uint32 {
	return InputInvoiceSlugTypeID
}

// TypeName returns name of type in TL schema.
func (*InputInvoiceSlug) TypeName() string {
	return "inputInvoiceSlug"
}

// TypeInfo returns info about TL type.
func (i *InputInvoiceSlug) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputInvoiceSlug",
		ID:   InputInvoiceSlugTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Slug",
			SchemaName: "slug",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputInvoiceSlug) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoiceSlug#c326caef as nil")
	}
	b.PutID(InputInvoiceSlugTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputInvoiceSlug) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoiceSlug#c326caef as nil")
	}
	b.PutString(i.Slug)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputInvoiceSlug) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoiceSlug#c326caef to nil")
	}
	if err := b.ConsumeID(InputInvoiceSlugTypeID); err != nil {
		return fmt.Errorf("unable to decode inputInvoiceSlug#c326caef: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputInvoiceSlug) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoiceSlug#c326caef to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode inputInvoiceSlug#c326caef: field slug: %w", err)
		}
		i.Slug = value
	}
	return nil
}

// GetSlug returns value of Slug field.
func (i *InputInvoiceSlug) GetSlug() (value string) {
	if i == nil {
		return
	}
	return i.Slug
}

// InputInvoicePremiumGiftCode represents TL type `inputInvoicePremiumGiftCode#98986c0d`.
//
// See https://core.telegram.org/constructor/inputInvoicePremiumGiftCode for reference.
type InputInvoicePremiumGiftCode struct {
	// Purpose field of InputInvoicePremiumGiftCode.
	Purpose InputStorePaymentPurposeClass
	// Option field of InputInvoicePremiumGiftCode.
	Option PremiumGiftCodeOption
}

// InputInvoicePremiumGiftCodeTypeID is TL type id of InputInvoicePremiumGiftCode.
const InputInvoicePremiumGiftCodeTypeID = 0x98986c0d

// construct implements constructor of InputInvoiceClass.
func (i InputInvoicePremiumGiftCode) construct() InputInvoiceClass { return &i }

// Ensuring interfaces in compile-time for InputInvoicePremiumGiftCode.
var (
	_ bin.Encoder     = &InputInvoicePremiumGiftCode{}
	_ bin.Decoder     = &InputInvoicePremiumGiftCode{}
	_ bin.BareEncoder = &InputInvoicePremiumGiftCode{}
	_ bin.BareDecoder = &InputInvoicePremiumGiftCode{}

	_ InputInvoiceClass = &InputInvoicePremiumGiftCode{}
)

func (i *InputInvoicePremiumGiftCode) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Purpose == nil) {
		return false
	}
	if !(i.Option.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputInvoicePremiumGiftCode) String() string {
	if i == nil {
		return "InputInvoicePremiumGiftCode(nil)"
	}
	type Alias InputInvoicePremiumGiftCode
	return fmt.Sprintf("InputInvoicePremiumGiftCode%+v", Alias(*i))
}

// FillFrom fills InputInvoicePremiumGiftCode from given interface.
func (i *InputInvoicePremiumGiftCode) FillFrom(from interface {
	GetPurpose() (value InputStorePaymentPurposeClass)
	GetOption() (value PremiumGiftCodeOption)
}) {
	i.Purpose = from.GetPurpose()
	i.Option = from.GetOption()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputInvoicePremiumGiftCode) TypeID() uint32 {
	return InputInvoicePremiumGiftCodeTypeID
}

// TypeName returns name of type in TL schema.
func (*InputInvoicePremiumGiftCode) TypeName() string {
	return "inputInvoicePremiumGiftCode"
}

// TypeInfo returns info about TL type.
func (i *InputInvoicePremiumGiftCode) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputInvoicePremiumGiftCode",
		ID:   InputInvoicePremiumGiftCodeTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Purpose",
			SchemaName: "purpose",
		},
		{
			Name:       "Option",
			SchemaName: "option",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputInvoicePremiumGiftCode) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoicePremiumGiftCode#98986c0d as nil")
	}
	b.PutID(InputInvoicePremiumGiftCodeTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputInvoicePremiumGiftCode) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputInvoicePremiumGiftCode#98986c0d as nil")
	}
	if i.Purpose == nil {
		return fmt.Errorf("unable to encode inputInvoicePremiumGiftCode#98986c0d: field purpose is nil")
	}
	if err := i.Purpose.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputInvoicePremiumGiftCode#98986c0d: field purpose: %w", err)
	}
	if err := i.Option.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputInvoicePremiumGiftCode#98986c0d: field option: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (i *InputInvoicePremiumGiftCode) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoicePremiumGiftCode#98986c0d to nil")
	}
	if err := b.ConsumeID(InputInvoicePremiumGiftCodeTypeID); err != nil {
		return fmt.Errorf("unable to decode inputInvoicePremiumGiftCode#98986c0d: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputInvoicePremiumGiftCode) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputInvoicePremiumGiftCode#98986c0d to nil")
	}
	{
		value, err := DecodeInputStorePaymentPurpose(b)
		if err != nil {
			return fmt.Errorf("unable to decode inputInvoicePremiumGiftCode#98986c0d: field purpose: %w", err)
		}
		i.Purpose = value
	}
	{
		if err := i.Option.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputInvoicePremiumGiftCode#98986c0d: field option: %w", err)
		}
	}
	return nil
}

// GetPurpose returns value of Purpose field.
func (i *InputInvoicePremiumGiftCode) GetPurpose() (value InputStorePaymentPurposeClass) {
	if i == nil {
		return
	}
	return i.Purpose
}

// GetOption returns value of Option field.
func (i *InputInvoicePremiumGiftCode) GetOption() (value PremiumGiftCodeOption) {
	if i == nil {
		return
	}
	return i.Option
}

// InputInvoiceClassName is schema name of InputInvoiceClass.
const InputInvoiceClassName = "InputInvoice"

// InputInvoiceClass represents InputInvoice generic type.
//
// See https://core.telegram.org/type/InputInvoice for reference.
//
// Example:
//
//	g, err := tg.DecodeInputInvoice(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tg.InputInvoiceMessage: // inputInvoiceMessage#c5b56859
//	case *tg.InputInvoiceSlug: // inputInvoiceSlug#c326caef
//	case *tg.InputInvoicePremiumGiftCode: // inputInvoicePremiumGiftCode#98986c0d
//	default: panic(v)
//	}
type InputInvoiceClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() InputInvoiceClass

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

// DecodeInputInvoice implements binary de-serialization for InputInvoiceClass.
func DecodeInputInvoice(buf *bin.Buffer) (InputInvoiceClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case InputInvoiceMessageTypeID:
		// Decoding inputInvoiceMessage#c5b56859.
		v := InputInvoiceMessage{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputInvoiceClass: %w", err)
		}
		return &v, nil
	case InputInvoiceSlugTypeID:
		// Decoding inputInvoiceSlug#c326caef.
		v := InputInvoiceSlug{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputInvoiceClass: %w", err)
		}
		return &v, nil
	case InputInvoicePremiumGiftCodeTypeID:
		// Decoding inputInvoicePremiumGiftCode#98986c0d.
		v := InputInvoicePremiumGiftCode{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputInvoiceClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode InputInvoiceClass: %w", bin.NewUnexpectedID(id))
	}
}

// InputInvoice boxes the InputInvoiceClass providing a helper.
type InputInvoiceBox struct {
	InputInvoice InputInvoiceClass
}

// Decode implements bin.Decoder for InputInvoiceBox.
func (b *InputInvoiceBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode InputInvoiceBox to nil")
	}
	v, err := DecodeInputInvoice(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.InputInvoice = v
	return nil
}

// Encode implements bin.Encode for InputInvoiceBox.
func (b *InputInvoiceBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.InputInvoice == nil {
		return fmt.Errorf("unable to encode InputInvoiceClass as nil")
	}
	return b.InputInvoice.Encode(buf)
}
