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

// PremiumGiveawayParticipantStatusEligible represents TL type `premiumGiveawayParticipantStatusEligible#7ee281c0`.
type PremiumGiveawayParticipantStatusEligible struct {
}

// PremiumGiveawayParticipantStatusEligibleTypeID is TL type id of PremiumGiveawayParticipantStatusEligible.
const PremiumGiveawayParticipantStatusEligibleTypeID = 0x7ee281c0

// construct implements constructor of PremiumGiveawayParticipantStatusClass.
func (p PremiumGiveawayParticipantStatusEligible) construct() PremiumGiveawayParticipantStatusClass {
	return &p
}

// Ensuring interfaces in compile-time for PremiumGiveawayParticipantStatusEligible.
var (
	_ bin.Encoder     = &PremiumGiveawayParticipantStatusEligible{}
	_ bin.Decoder     = &PremiumGiveawayParticipantStatusEligible{}
	_ bin.BareEncoder = &PremiumGiveawayParticipantStatusEligible{}
	_ bin.BareDecoder = &PremiumGiveawayParticipantStatusEligible{}

	_ PremiumGiveawayParticipantStatusClass = &PremiumGiveawayParticipantStatusEligible{}
)

func (p *PremiumGiveawayParticipantStatusEligible) Zero() bool {
	if p == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayParticipantStatusEligible) String() string {
	if p == nil {
		return "PremiumGiveawayParticipantStatusEligible(nil)"
	}
	type Alias PremiumGiveawayParticipantStatusEligible
	return fmt.Sprintf("PremiumGiveawayParticipantStatusEligible%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayParticipantStatusEligible) TypeID() uint32 {
	return PremiumGiveawayParticipantStatusEligibleTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayParticipantStatusEligible) TypeName() string {
	return "premiumGiveawayParticipantStatusEligible"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayParticipantStatusEligible) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayParticipantStatusEligible",
		ID:   PremiumGiveawayParticipantStatusEligibleTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayParticipantStatusEligible) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusEligible#7ee281c0 as nil")
	}
	b.PutID(PremiumGiveawayParticipantStatusEligibleTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayParticipantStatusEligible) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusEligible#7ee281c0 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayParticipantStatusEligible) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusEligible#7ee281c0 to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayParticipantStatusEligibleTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusEligible#7ee281c0: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayParticipantStatusEligible) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusEligible#7ee281c0 to nil")
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayParticipantStatusEligible) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusEligible#7ee281c0 as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayParticipantStatusEligible")
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayParticipantStatusEligible) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusEligible#7ee281c0 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayParticipantStatusEligible"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusEligible#7ee281c0: %w", err)
			}
		default:
			return b.Skip()
		}
		return nil
	})
}

// PremiumGiveawayParticipantStatusParticipating represents TL type `premiumGiveawayParticipantStatusParticipating#740406d1`.
type PremiumGiveawayParticipantStatusParticipating struct {
}

// PremiumGiveawayParticipantStatusParticipatingTypeID is TL type id of PremiumGiveawayParticipantStatusParticipating.
const PremiumGiveawayParticipantStatusParticipatingTypeID = 0x740406d1

// construct implements constructor of PremiumGiveawayParticipantStatusClass.
func (p PremiumGiveawayParticipantStatusParticipating) construct() PremiumGiveawayParticipantStatusClass {
	return &p
}

// Ensuring interfaces in compile-time for PremiumGiveawayParticipantStatusParticipating.
var (
	_ bin.Encoder     = &PremiumGiveawayParticipantStatusParticipating{}
	_ bin.Decoder     = &PremiumGiveawayParticipantStatusParticipating{}
	_ bin.BareEncoder = &PremiumGiveawayParticipantStatusParticipating{}
	_ bin.BareDecoder = &PremiumGiveawayParticipantStatusParticipating{}

	_ PremiumGiveawayParticipantStatusClass = &PremiumGiveawayParticipantStatusParticipating{}
)

func (p *PremiumGiveawayParticipantStatusParticipating) Zero() bool {
	if p == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayParticipantStatusParticipating) String() string {
	if p == nil {
		return "PremiumGiveawayParticipantStatusParticipating(nil)"
	}
	type Alias PremiumGiveawayParticipantStatusParticipating
	return fmt.Sprintf("PremiumGiveawayParticipantStatusParticipating%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayParticipantStatusParticipating) TypeID() uint32 {
	return PremiumGiveawayParticipantStatusParticipatingTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayParticipantStatusParticipating) TypeName() string {
	return "premiumGiveawayParticipantStatusParticipating"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayParticipantStatusParticipating) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayParticipantStatusParticipating",
		ID:   PremiumGiveawayParticipantStatusParticipatingTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayParticipantStatusParticipating) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusParticipating#740406d1 as nil")
	}
	b.PutID(PremiumGiveawayParticipantStatusParticipatingTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayParticipantStatusParticipating) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusParticipating#740406d1 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayParticipantStatusParticipating) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusParticipating#740406d1 to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayParticipantStatusParticipatingTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusParticipating#740406d1: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayParticipantStatusParticipating) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusParticipating#740406d1 to nil")
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayParticipantStatusParticipating) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusParticipating#740406d1 as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayParticipantStatusParticipating")
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayParticipantStatusParticipating) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusParticipating#740406d1 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayParticipantStatusParticipating"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusParticipating#740406d1: %w", err)
			}
		default:
			return b.Skip()
		}
		return nil
	})
}

// PremiumGiveawayParticipantStatusAlreadyWasMember represents TL type `premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3`.
type PremiumGiveawayParticipantStatusAlreadyWasMember struct {
	// Point in time (Unix timestamp) when the user joined the chat
	JoinedChatDate int32
}

// PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID is TL type id of PremiumGiveawayParticipantStatusAlreadyWasMember.
const PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID = 0x8d3045a3

// construct implements constructor of PremiumGiveawayParticipantStatusClass.
func (p PremiumGiveawayParticipantStatusAlreadyWasMember) construct() PremiumGiveawayParticipantStatusClass {
	return &p
}

// Ensuring interfaces in compile-time for PremiumGiveawayParticipantStatusAlreadyWasMember.
var (
	_ bin.Encoder     = &PremiumGiveawayParticipantStatusAlreadyWasMember{}
	_ bin.Decoder     = &PremiumGiveawayParticipantStatusAlreadyWasMember{}
	_ bin.BareEncoder = &PremiumGiveawayParticipantStatusAlreadyWasMember{}
	_ bin.BareDecoder = &PremiumGiveawayParticipantStatusAlreadyWasMember{}

	_ PremiumGiveawayParticipantStatusClass = &PremiumGiveawayParticipantStatusAlreadyWasMember{}
)

func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.JoinedChatDate == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) String() string {
	if p == nil {
		return "PremiumGiveawayParticipantStatusAlreadyWasMember(nil)"
	}
	type Alias PremiumGiveawayParticipantStatusAlreadyWasMember
	return fmt.Sprintf("PremiumGiveawayParticipantStatusAlreadyWasMember%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayParticipantStatusAlreadyWasMember) TypeID() uint32 {
	return PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayParticipantStatusAlreadyWasMember) TypeName() string {
	return "premiumGiveawayParticipantStatusAlreadyWasMember"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayParticipantStatusAlreadyWasMember",
		ID:   PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "JoinedChatDate",
			SchemaName: "joined_chat_date",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 as nil")
	}
	b.PutID(PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 as nil")
	}
	b.PutInt32(p.JoinedChatDate)
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 to nil")
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3: field joined_chat_date: %w", err)
		}
		p.JoinedChatDate = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayParticipantStatusAlreadyWasMember")
	b.Comma()
	b.FieldStart("joined_chat_date")
	b.PutInt32(p.JoinedChatDate)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayParticipantStatusAlreadyWasMember"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3: %w", err)
			}
		case "joined_chat_date":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3: field joined_chat_date: %w", err)
			}
			p.JoinedChatDate = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetJoinedChatDate returns value of JoinedChatDate field.
func (p *PremiumGiveawayParticipantStatusAlreadyWasMember) GetJoinedChatDate() (value int32) {
	if p == nil {
		return
	}
	return p.JoinedChatDate
}

// PremiumGiveawayParticipantStatusAdministrator represents TL type `premiumGiveawayParticipantStatusAdministrator#7244dace`.
type PremiumGiveawayParticipantStatusAdministrator struct {
	// Identifier of the chat administered by the user
	ChatID int64
}

// PremiumGiveawayParticipantStatusAdministratorTypeID is TL type id of PremiumGiveawayParticipantStatusAdministrator.
const PremiumGiveawayParticipantStatusAdministratorTypeID = 0x7244dace

// construct implements constructor of PremiumGiveawayParticipantStatusClass.
func (p PremiumGiveawayParticipantStatusAdministrator) construct() PremiumGiveawayParticipantStatusClass {
	return &p
}

// Ensuring interfaces in compile-time for PremiumGiveawayParticipantStatusAdministrator.
var (
	_ bin.Encoder     = &PremiumGiveawayParticipantStatusAdministrator{}
	_ bin.Decoder     = &PremiumGiveawayParticipantStatusAdministrator{}
	_ bin.BareEncoder = &PremiumGiveawayParticipantStatusAdministrator{}
	_ bin.BareDecoder = &PremiumGiveawayParticipantStatusAdministrator{}

	_ PremiumGiveawayParticipantStatusClass = &PremiumGiveawayParticipantStatusAdministrator{}
)

func (p *PremiumGiveawayParticipantStatusAdministrator) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.ChatID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayParticipantStatusAdministrator) String() string {
	if p == nil {
		return "PremiumGiveawayParticipantStatusAdministrator(nil)"
	}
	type Alias PremiumGiveawayParticipantStatusAdministrator
	return fmt.Sprintf("PremiumGiveawayParticipantStatusAdministrator%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayParticipantStatusAdministrator) TypeID() uint32 {
	return PremiumGiveawayParticipantStatusAdministratorTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayParticipantStatusAdministrator) TypeName() string {
	return "premiumGiveawayParticipantStatusAdministrator"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayParticipantStatusAdministrator) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayParticipantStatusAdministrator",
		ID:   PremiumGiveawayParticipantStatusAdministratorTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ChatID",
			SchemaName: "chat_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAdministrator#7244dace as nil")
	}
	b.PutID(PremiumGiveawayParticipantStatusAdministratorTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAdministrator#7244dace as nil")
	}
	b.PutInt53(p.ChatID)
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAdministrator#7244dace to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayParticipantStatusAdministratorTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAdministrator#7244dace: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAdministrator#7244dace to nil")
	}
	{
		value, err := b.Int53()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAdministrator#7244dace: field chat_id: %w", err)
		}
		p.ChatID = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusAdministrator#7244dace as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayParticipantStatusAdministrator")
	b.Comma()
	b.FieldStart("chat_id")
	b.PutInt53(p.ChatID)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayParticipantStatusAdministrator) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusAdministrator#7244dace to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayParticipantStatusAdministrator"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAdministrator#7244dace: %w", err)
			}
		case "chat_id":
			value, err := b.Int53()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusAdministrator#7244dace: field chat_id: %w", err)
			}
			p.ChatID = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetChatID returns value of ChatID field.
func (p *PremiumGiveawayParticipantStatusAdministrator) GetChatID() (value int64) {
	if p == nil {
		return
	}
	return p.ChatID
}

// PremiumGiveawayParticipantStatusDisallowedCountry represents TL type `premiumGiveawayParticipantStatusDisallowedCountry#89684090`.
type PremiumGiveawayParticipantStatusDisallowedCountry struct {
	// A two-letter ISO 3166-1 alpha-2 country code of the user's country
	UserCountryCode string
}

// PremiumGiveawayParticipantStatusDisallowedCountryTypeID is TL type id of PremiumGiveawayParticipantStatusDisallowedCountry.
const PremiumGiveawayParticipantStatusDisallowedCountryTypeID = 0x89684090

// construct implements constructor of PremiumGiveawayParticipantStatusClass.
func (p PremiumGiveawayParticipantStatusDisallowedCountry) construct() PremiumGiveawayParticipantStatusClass {
	return &p
}

// Ensuring interfaces in compile-time for PremiumGiveawayParticipantStatusDisallowedCountry.
var (
	_ bin.Encoder     = &PremiumGiveawayParticipantStatusDisallowedCountry{}
	_ bin.Decoder     = &PremiumGiveawayParticipantStatusDisallowedCountry{}
	_ bin.BareEncoder = &PremiumGiveawayParticipantStatusDisallowedCountry{}
	_ bin.BareDecoder = &PremiumGiveawayParticipantStatusDisallowedCountry{}

	_ PremiumGiveawayParticipantStatusClass = &PremiumGiveawayParticipantStatusDisallowedCountry{}
)

func (p *PremiumGiveawayParticipantStatusDisallowedCountry) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.UserCountryCode == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) String() string {
	if p == nil {
		return "PremiumGiveawayParticipantStatusDisallowedCountry(nil)"
	}
	type Alias PremiumGiveawayParticipantStatusDisallowedCountry
	return fmt.Sprintf("PremiumGiveawayParticipantStatusDisallowedCountry%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGiveawayParticipantStatusDisallowedCountry) TypeID() uint32 {
	return PremiumGiveawayParticipantStatusDisallowedCountryTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGiveawayParticipantStatusDisallowedCountry) TypeName() string {
	return "premiumGiveawayParticipantStatusDisallowedCountry"
}

// TypeInfo returns info about TL type.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premiumGiveawayParticipantStatusDisallowedCountry",
		ID:   PremiumGiveawayParticipantStatusDisallowedCountryTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "UserCountryCode",
			SchemaName: "user_country_code",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusDisallowedCountry#89684090 as nil")
	}
	b.PutID(PremiumGiveawayParticipantStatusDisallowedCountryTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusDisallowedCountry#89684090 as nil")
	}
	b.PutString(p.UserCountryCode)
	return nil
}

// Decode implements bin.Decoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusDisallowedCountry#89684090 to nil")
	}
	if err := b.ConsumeID(PremiumGiveawayParticipantStatusDisallowedCountryTypeID); err != nil {
		return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusDisallowedCountry#89684090: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusDisallowedCountry#89684090 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusDisallowedCountry#89684090: field user_country_code: %w", err)
		}
		p.UserCountryCode = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) EncodeTDLibJSON(b tdjson.Encoder) error {
	if p == nil {
		return fmt.Errorf("can't encode premiumGiveawayParticipantStatusDisallowedCountry#89684090 as nil")
	}
	b.ObjStart()
	b.PutID("premiumGiveawayParticipantStatusDisallowedCountry")
	b.Comma()
	b.FieldStart("user_country_code")
	b.PutString(p.UserCountryCode)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) DecodeTDLibJSON(b tdjson.Decoder) error {
	if p == nil {
		return fmt.Errorf("can't decode premiumGiveawayParticipantStatusDisallowedCountry#89684090 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("premiumGiveawayParticipantStatusDisallowedCountry"); err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusDisallowedCountry#89684090: %w", err)
			}
		case "user_country_code":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode premiumGiveawayParticipantStatusDisallowedCountry#89684090: field user_country_code: %w", err)
			}
			p.UserCountryCode = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetUserCountryCode returns value of UserCountryCode field.
func (p *PremiumGiveawayParticipantStatusDisallowedCountry) GetUserCountryCode() (value string) {
	if p == nil {
		return
	}
	return p.UserCountryCode
}

// PremiumGiveawayParticipantStatusClassName is schema name of PremiumGiveawayParticipantStatusClass.
const PremiumGiveawayParticipantStatusClassName = "PremiumGiveawayParticipantStatus"

// PremiumGiveawayParticipantStatusClass represents PremiumGiveawayParticipantStatus generic type.
//
// Example:
//
//	g, err := tdapi.DecodePremiumGiveawayParticipantStatus(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tdapi.PremiumGiveawayParticipantStatusEligible: // premiumGiveawayParticipantStatusEligible#7ee281c0
//	case *tdapi.PremiumGiveawayParticipantStatusParticipating: // premiumGiveawayParticipantStatusParticipating#740406d1
//	case *tdapi.PremiumGiveawayParticipantStatusAlreadyWasMember: // premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3
//	case *tdapi.PremiumGiveawayParticipantStatusAdministrator: // premiumGiveawayParticipantStatusAdministrator#7244dace
//	case *tdapi.PremiumGiveawayParticipantStatusDisallowedCountry: // premiumGiveawayParticipantStatusDisallowedCountry#89684090
//	default: panic(v)
//	}
type PremiumGiveawayParticipantStatusClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() PremiumGiveawayParticipantStatusClass

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
}

// DecodePremiumGiveawayParticipantStatus implements binary de-serialization for PremiumGiveawayParticipantStatusClass.
func DecodePremiumGiveawayParticipantStatus(buf *bin.Buffer) (PremiumGiveawayParticipantStatusClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case PremiumGiveawayParticipantStatusEligibleTypeID:
		// Decoding premiumGiveawayParticipantStatusEligible#7ee281c0.
		v := PremiumGiveawayParticipantStatusEligible{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case PremiumGiveawayParticipantStatusParticipatingTypeID:
		// Decoding premiumGiveawayParticipantStatusParticipating#740406d1.
		v := PremiumGiveawayParticipantStatusParticipating{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case PremiumGiveawayParticipantStatusAlreadyWasMemberTypeID:
		// Decoding premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3.
		v := PremiumGiveawayParticipantStatusAlreadyWasMember{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case PremiumGiveawayParticipantStatusAdministratorTypeID:
		// Decoding premiumGiveawayParticipantStatusAdministrator#7244dace.
		v := PremiumGiveawayParticipantStatusAdministrator{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case PremiumGiveawayParticipantStatusDisallowedCountryTypeID:
		// Decoding premiumGiveawayParticipantStatusDisallowedCountry#89684090.
		v := PremiumGiveawayParticipantStatusDisallowedCountry{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", bin.NewUnexpectedID(id))
	}
}

// DecodeTDLibJSONPremiumGiveawayParticipantStatus implements binary de-serialization for PremiumGiveawayParticipantStatusClass.
func DecodeTDLibJSONPremiumGiveawayParticipantStatus(buf tdjson.Decoder) (PremiumGiveawayParticipantStatusClass, error) {
	id, err := buf.FindTypeID()
	if err != nil {
		return nil, err
	}
	switch id {
	case "premiumGiveawayParticipantStatusEligible":
		// Decoding premiumGiveawayParticipantStatusEligible#7ee281c0.
		v := PremiumGiveawayParticipantStatusEligible{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case "premiumGiveawayParticipantStatusParticipating":
		// Decoding premiumGiveawayParticipantStatusParticipating#740406d1.
		v := PremiumGiveawayParticipantStatusParticipating{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case "premiumGiveawayParticipantStatusAlreadyWasMember":
		// Decoding premiumGiveawayParticipantStatusAlreadyWasMember#8d3045a3.
		v := PremiumGiveawayParticipantStatusAlreadyWasMember{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case "premiumGiveawayParticipantStatusAdministrator":
		// Decoding premiumGiveawayParticipantStatusAdministrator#7244dace.
		v := PremiumGiveawayParticipantStatusAdministrator{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	case "premiumGiveawayParticipantStatusDisallowedCountry":
		// Decoding premiumGiveawayParticipantStatusDisallowedCountry#89684090.
		v := PremiumGiveawayParticipantStatusDisallowedCountry{}
		if err := v.DecodeTDLibJSON(buf); err != nil {
			return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusClass: %w", tdjson.NewUnexpectedID(id))
	}
}

// PremiumGiveawayParticipantStatus boxes the PremiumGiveawayParticipantStatusClass providing a helper.
type PremiumGiveawayParticipantStatusBox struct {
	PremiumGiveawayParticipantStatus PremiumGiveawayParticipantStatusClass
}

// Decode implements bin.Decoder for PremiumGiveawayParticipantStatusBox.
func (b *PremiumGiveawayParticipantStatusBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusBox to nil")
	}
	v, err := DecodePremiumGiveawayParticipantStatus(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PremiumGiveawayParticipantStatus = v
	return nil
}

// Encode implements bin.Encode for PremiumGiveawayParticipantStatusBox.
func (b *PremiumGiveawayParticipantStatusBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.PremiumGiveawayParticipantStatus == nil {
		return fmt.Errorf("unable to encode PremiumGiveawayParticipantStatusClass as nil")
	}
	return b.PremiumGiveawayParticipantStatus.Encode(buf)
}

// DecodeTDLibJSON implements bin.Decoder for PremiumGiveawayParticipantStatusBox.
func (b *PremiumGiveawayParticipantStatusBox) DecodeTDLibJSON(buf tdjson.Decoder) error {
	if b == nil {
		return fmt.Errorf("unable to decode PremiumGiveawayParticipantStatusBox to nil")
	}
	v, err := DecodeTDLibJSONPremiumGiveawayParticipantStatus(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PremiumGiveawayParticipantStatus = v
	return nil
}

// EncodeTDLibJSON implements bin.Encode for PremiumGiveawayParticipantStatusBox.
func (b *PremiumGiveawayParticipantStatusBox) EncodeTDLibJSON(buf tdjson.Encoder) error {
	if b == nil || b.PremiumGiveawayParticipantStatus == nil {
		return fmt.Errorf("unable to encode PremiumGiveawayParticipantStatusClass as nil")
	}
	return b.PremiumGiveawayParticipantStatus.EncodeTDLibJSON(buf)
}
