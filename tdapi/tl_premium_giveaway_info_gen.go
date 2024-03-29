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

// PremiumGiveawayInfoOngoing represents TL type `premiumGiveawayInfoOngoing#48d3ce6f`.
type PremiumGiveawayInfoOngoing struct {
	// Point in time (Unix timestamp) when the giveaway was created
	CreationDate int32
	// Status of the current user in the giveaway
	Status PremiumGiveawayParticipantStatusClass
	// True, if the giveaway has ended and results are being prepared
	IsEnded bool
}

// PremiumGiveawayInfoOngoingTypeID is TL type id of PremiumGiveawayInfoOngoing.
const PremiumGiveawayInfoOngoingTypeID = 0x48d3ce6f

// construct implements constructor of PremiumGiveawayInfoClass.
func (p PremiumGiveawayInfoOngoing) construct() PremiumGiveawayInfoClass { return &p }

// Ensuring interfaces in compile-time for PremiumGiveawayInfoOngoing.
var (
	_ bin.Encoder     = &PremiumGiveawayInfoOngoing{}
	_ bin.Decoder     = &PremiumGiveawayInfoOngoing{}
	_ bin.BareEncoder = &PremiumGiveawayInfoOngoing{}
	_ bin.BareDecoder = &PremiumGiveawayInfoOngoing{}

	_ PremiumGiveawayInfoClass = &PremiumGiveawayInfoOngoing{}
)

func (p *PremiumGiveawayInfoOngoing) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.CreationDate == 0) {
		return false
	}
	if !(p.Status == nil) {
		return false
	}
	if !(p.IsEnded == false) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayInfoOngoing) String() string {
	if p == nil {
		return "PremiumGiveawayInfoOngoing(nil)"
	}
	type Alias PremiumGiveawayInfoOngoing
	return fmt.Sprintf("PremiumGiveawayInfoOngoing%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayInfoOngoing) TypeID() uint32 {
	return PremiumGiveawayInfoOngoingTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayInfoOngoing) TypeName() string {
	return "premiumGiveawayInfoOngoing"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayInfoOngoing) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayInfoOngoing",
		ID:   PremiumGiveawayInfoOngoingTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "CreationDate",
			SchemaName: "creation_date",
		},
		{
			Name:       "Status",
			SchemaName: "status",
		},
		{
			Name:       "IsEnded",
			SchemaName: "is_ended",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayInfoOngoing) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoOngoing#48d3ce6f as nil")
	}
	b.PutID(PremiumGiveawayInfoOngoingTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayInfoOngoing) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoOngoing#48d3ce6f as nil")
	}
	b.PutInt32(p.CreationDate)
	if p.Status == nil {
		return fmt.Errorf("unable to encode premiumGiveawayInfoOngoing#48d3ce6f: field status is nil")
	}
	if err := p.Status.Encode(b); err != nil {
		return fmt.Errorf("unable to encode premiumGiveawayInfoOngoing#48d3ce6f: field status: %w", err)
	}
	b.PutBool(p.IsEnded)
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayInfoOngoing) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoOngoing#48d3ce6f to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayInfoOngoingTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayInfoOngoing) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoOngoing#48d3ce6f to nil")
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field creation_date: %w", err)
		}
		p.CreationDate = value
	}
	{
		value, err := DecodePremiumGiveawayParticipantStatus(b)
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field status: %w", err)
		}
		p.Status = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field is_ended: %w", err)
		}
		p.IsEnded = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayInfoOngoing) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoOngoing#48d3ce6f as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayInfoOngoing")
	b.Comma()
	b.FieldStart("creation_date")
	b.PutInt32(p.CreationDate)
	b.Comma()
	b.FieldStart("status")
	if p.Status == nil {
		return fmt.Errorf("unable to encode premiumGiveawayInfoOngoing#48d3ce6f: field status is nil")
	}
	if err := p.Status.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode premiumGiveawayInfoOngoing#48d3ce6f: field status: %w", err)
	}
	b.Comma()
	b.FieldStart("is_ended")
	b.PutBool(p.IsEnded)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayInfoOngoing) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoOngoing#48d3ce6f to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayInfoOngoing"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: %w", err)
			}
		case "creation_date":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field creation_date: %w", err)
			}
			p.CreationDate = value
		case "status":
			value, err := DecodeTDLibJSONPremiumGiveawayParticipantStatus(b)
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field status: %w", err)
			}
			p.Status = value
		case "is_ended":
			value, err := b.Bool()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoOngoing#48d3ce6f: field is_ended: %w", err)
			}
			p.IsEnded = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetCreationDate returns value of CreationDate field.
func (p *PremiumGiveawayInfoOngoing) GetCreationDate() (value int32) {
	if p == nil {
		return
	}
	return p.CreationDate
}

// GetStatus returns value of Status field.
func (p *PremiumGiveawayInfoOngoing) GetStatus() (value PremiumGiveawayParticipantStatusClass) {
	if p == nil {
		return
	}
	return p.Status
}

// GetIsEnded returns value of IsEnded field.
func (p *PremiumGiveawayInfoOngoing) GetIsEnded() (value bool) {
	if p == nil {
		return
	}
	return p.IsEnded
}

// PremiumGiveawayInfoCompleted represents TL type `premiumGiveawayInfoCompleted#fc8b501b`.
type PremiumGiveawayInfoCompleted struct {
	// Point in time (Unix timestamp) when the giveaway was created
	CreationDate int32
	// Point in time (Unix timestamp) when the winners were selected. May be bigger than
	// winners selection date specified in parameters of the giveaway
	ActualWinnersSelectionDate int32
	// True, if the giveaway was canceled and was fully refunded
	WasRefunded bool
	// Number of winners in the giveaway
	WinnerCount int32
	// Number of winners, which activated their gift codes
	ActivationCount int32
	// Telegram Premium gift code that was received by the current user; empty if the user
	// isn't a winner in the giveaway
	GiftCode string
}

// PremiumGiveawayInfoCompletedTypeID is TL type id of PremiumGiveawayInfoCompleted.
const PremiumGiveawayInfoCompletedTypeID = 0xfc8b501b

// construct implements constructor of PremiumGiveawayInfoClass.
func (p PremiumGiveawayInfoCompleted) construct() PremiumGiveawayInfoClass { return &p }

// Ensuring interfaces in compile-time for PremiumGiveawayInfoCompleted.
var (
	_ bin.Encoder     = &PremiumGiveawayInfoCompleted{}
	_ bin.Decoder     = &PremiumGiveawayInfoCompleted{}
	_ bin.BareEncoder = &PremiumGiveawayInfoCompleted{}
	_ bin.BareDecoder = &PremiumGiveawayInfoCompleted{}

	_ PremiumGiveawayInfoClass = &PremiumGiveawayInfoCompleted{}
)

func (p *PremiumGiveawayInfoCompleted) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.CreationDate == 0) {
		return false
	}
	if !(p.ActualWinnersSelectionDate == 0) {
		return false
	}
	if !(p.WasRefunded == false) {
		return false
	}
	if !(p.WinnerCount == 0) {
		return false
	}
	if !(p.ActivationCount == 0) {
		return false
	}
	if !(p.GiftCode == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayInfoCompleted) String() string {
	if p == nil {
		return "PremiumGiveawayInfoCompleted(nil)"
	}
	type Alias PremiumGiveawayInfoCompleted
	return fmt.Sprintf("PremiumGiveawayInfoCompleted%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayInfoCompleted) TypeID() uint32 {
	return PremiumGiveawayInfoCompletedTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayInfoCompleted) TypeName() string {
	return "premiumGiveawayInfoCompleted"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayInfoCompleted) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayInfoCompleted",
		ID:   PremiumGiveawayInfoCompletedTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "CreationDate",
			SchemaName: "creation_date",
		},
		{
			Name:       "ActualWinnersSelectionDate",
			SchemaName: "actual_winners_selection_date",
		},
		{
			Name:       "WasRefunded",
			SchemaName: "was_refunded",
		},
		{
			Name:       "WinnerCount",
			SchemaName: "winner_count",
		},
		{
			Name:       "ActivationCount",
			SchemaName: "activation_count",
		},
		{
			Name:       "GiftCode",
			SchemaName: "gift_code",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayInfoCompleted) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoCompleted#fc8b501b as nil")
	}
	b.PutID(PremiumGiveawayInfoCompletedTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayInfoCompleted) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoCompleted#fc8b501b as nil")
	}
	b.PutInt32(p.CreationDate)
	b.PutInt32(p.ActualWinnersSelectionDate)
	b.PutBool(p.WasRefunded)
	b.PutInt32(p.WinnerCount)
	b.PutInt32(p.ActivationCount)
	b.PutString(p.GiftCode)
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayInfoCompleted) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoCompleted#fc8b501b to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayInfoCompletedTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayInfoCompleted) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoCompleted#fc8b501b to nil")
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field creation_date: %w", err)
		}
		p.CreationDate = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field actual_winners_selection_date: %w", err)
		}
		p.ActualWinnersSelectionDate = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field was_refunded: %w", err)
		}
		p.WasRefunded = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field winner_count: %w", err)
		}
		p.WinnerCount = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field activation_count: %w", err)
		}
		p.ActivationCount = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field gift_code: %w", err)
		}
		p.GiftCode = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayInfoCompleted) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayInfoCompleted#fc8b501b as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayInfoCompleted")
	b.Comma()
	b.FieldStart("creation_date")
	b.PutInt32(p.CreationDate)
	b.Comma()
	b.FieldStart("actual_winners_selection_date")
	b.PutInt32(p.ActualWinnersSelectionDate)
	b.Comma()
	b.FieldStart("was_refunded")
	b.PutBool(p.WasRefunded)
	b.Comma()
	b.FieldStart("winner_count")
	b.PutInt32(p.WinnerCount)
	b.Comma()
	b.FieldStart("activation_count")
	b.PutInt32(p.ActivationCount)
	b.Comma()
	b.FieldStart("gift_code")
	b.PutString(p.GiftCode)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayInfoCompleted) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayInfoCompleted#fc8b501b to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayInfoCompleted"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: %w", err)
			}
		case "creation_date":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field creation_date: %w", err)
			}
			p.CreationDate = value
		case "actual_winners_selection_date":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field actual_winners_selection_date: %w", err)
			}
			p.ActualWinnersSelectionDate = value
		case "was_refunded":
			value, err := b.Bool()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field was_refunded: %w", err)
			}
			p.WasRefunded = value
		case "winner_count":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field winner_count: %w", err)
			}
			p.WinnerCount = value
		case "activation_count":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field activation_count: %w", err)
			}
			p.ActivationCount = value
		case "gift_code":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayInfoCompleted#fc8b501b: field gift_code: %w", err)
			}
			p.GiftCode = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetCreationDate returns value of CreationDate field.
func (p *PremiumGiveawayInfoCompleted) GetCreationDate() (value int32) {
	if p == nil {
		return
	}
	return p.CreationDate
}

// GetActualWinnersSelectionDate returns value of ActualWinnersSelectionDate field.
func (p *PremiumGiveawayInfoCompleted) GetActualWinnersSelectionDate() (value int32) {
	if p == nil {
		return
	}
	return p.ActualWinnersSelectionDate
}

// GetWasRefunded returns value of WasRefunded field.
func (p *PremiumGiveawayInfoCompleted) GetWasRefunded() (value bool) {
	if p == nil {
		return
	}
	return p.WasRefunded
}

// GetWinnerCount returns value of WinnerCount field.
func (p *PremiumGiveawayInfoCompleted) GetWinnerCount() (value int32) {
	if p == nil {
		return
	}
	return p.WinnerCount
}

// GetActivationCount returns value of ActivationCount field.
func (p *PremiumGiveawayInfoCompleted) GetActivationCount() (value int32) {
	if p == nil {
		return
	}
	return p.ActivationCount
}

// GetGiftCode returns value of GiftCode field.
func (p *PremiumGiveawayInfoCompleted) GetGiftCode() (value string) {
	if p == nil {
		return
	}
	return p.GiftCode
}

// PremiumGiveawayInfoClassName is schema name of PremiumGiveawayInfoClass.
const PremiumGiveawayInfoClassName = "PremiumGiveawayInfo"

// PremiumGiveawayInfoClass represents PremiumGiveawayInfo generic type.
//
// Example:
//
//	g, err := tdapi.DecodePremiumGiveawayInfo(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tdapi.PremiumGiveawayInfoOngoing: // premiumGiveawayInfoOngoing#48d3ce6f
//	case *tdapi.PremiumGiveawayInfoCompleted: // premiumGiveawayInfoCompleted#fc8b501b
//	default: panic(v)
//	}
type PremiumGiveawayInfoClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() PremiumGiveawayInfoClass

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

	EncodeTDLibJSON(b tdjson.Encoder) error
	DecodeTDLibJSON(b tdjson.Decoder) error

	// Point in time (Unix timestamp) when the giveaway was created
	GetCreationDate() (value int32)
}

// DecodePremiumGiveawayInfo implements binary de-serialization for PremiumGiveawayInfoClass.
func DecodePremiumGiveawayInfo(buf *bin.Buffer) (PremiumGiveawayInfoClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case PremiumGiveawayInfoOngoingTypeID:
		// Decoding premiumGiveawayInfoOngoing#48d3ce6f.
		v := PremiumGiveawayInfoOngoing{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", err)
		}
		return &v, nil
	case PremiumGiveawayInfoCompletedTypeID:
		// Decoding premiumGiveawayInfoCompleted#fc8b501b.
		v := PremiumGiveawayInfoCompleted{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", bin.NewUnexpectedID(id))
	}
}

// DecodeTDLibJSONPremiumGiveawayInfo implements binary de-serialization for PremiumGiveawayInfoClass.
func DecodeTDLibJSONPremiumGiveawayInfo(buf tdjson.Decoder) (PremiumGiveawayInfoClass, error) {
	id, err := buf.FindTypeID()
	if err != nil {
		return nil, err
	}
	switch id {
	case "premiumGiveawayInfoOngoing":
		// Decoding premiumGiveawayInfoOngoing#48d3ce6f.
		v := PremiumGiveawayInfoOngoing{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", err)
		}
		return &v, nil
	case "premiumGiveawayInfoCompleted":
		// Decoding premiumGiveawayInfoCompleted#fc8b501b.
		v := PremiumGiveawayInfoCompleted{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode PremiumGiveawayInfoClass: %w", tdjson.NewUnexpectedID(id))
	}
}

// PremiumGiveawayInfo boxes the PremiumGiveawayInfoClass providing a helper.
type PremiumGiveawayInfoBox struct {
	PremiumGiveawayInfo PremiumGiveawayInfoClass
}

// Decode implements bin.Decoder for PremiumGiveawayInfoBox.
func (b *PremiumGiveawayInfoBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode PremiumGiveawayInfoBox to nil")
	}
	v, err := DecodePremiumGiveawayInfo(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PremiumGiveawayInfo = v
	return nil
}

// Encode implements bin.Encode for PremiumGiveawayInfoBox.
func (b *PremiumGiveawayInfoBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.PremiumGiveawayInfo == nil {
		return fmt.Errorf("unable to encode PremiumGiveawayInfoClass as nil")
	}
	return b.PremiumGiveawayInfo.Encode(buf)
}

// DecodeTDLibJSON implements bin.Decoder for PremiumGiveawayInfoBox.
func (b *PremiumGiveawayInfoBox) DecodeTDLibJSON(buf tdjson.Decoder) error {
	if b == nil {
		return fmt.Errorf("unable to decode PremiumGiveawayInfoBox to nil")
	}
	v, err := DecodeTDLibJSONPremiumGiveawayInfo(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PremiumGiveawayInfo = v
	return nil
}

// EncodeTDLibJSON implements bin.Encode for PremiumGiveawayInfoBox.
func (b *PremiumGiveawayInfoBox) EncodeTDLibJSON(buf tdjson.Encoder) error {
	if b == nil || b.PremiumGiveawayInfo == nil {
		return fmt.Errorf("unable to encode PremiumGiveawayInfoClass as nil")
	}
	return b.PremiumGiveawayInfo.EncodeTDLibJSON(buf)
}
