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

// InputStorePaymentPremiumSubscription represents TL type `inputStorePaymentPremiumSubscription#a6751e66`.
// Info about a Telegram Premium purchase
//
// See https://core.telegram.org/constructor/inputStorePaymentPremiumSubscription for reference.
type InputStorePaymentPremiumSubscription struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Pass true if this is a restore of a Telegram Premium purchase; only for the App Store
	Restore bool
	// Pass true if this is an upgrade from a monthly subscription to a yearly subscription;
	// only for App Store
	Upgrade bool
}

// InputStorePaymentPremiumSubscriptionTypeID is TL type id of InputStorePaymentPremiumSubscription.
const InputStorePaymentPremiumSubscriptionTypeID = 0xa6751e66

// construct implements constructor of InputStorePaymentPurposeClass.
func (i InputStorePaymentPremiumSubscription) construct() InputStorePaymentPurposeClass { return &i }

// Ensuring interfaces in compile-time for InputStorePaymentPremiumSubscription.
var (
	_ bin.Encoder     = &InputStorePaymentPremiumSubscription{}
	_ bin.Decoder     = &InputStorePaymentPremiumSubscription{}
	_ bin.BareEncoder = &InputStorePaymentPremiumSubscription{}
	_ bin.BareDecoder = &InputStorePaymentPremiumSubscription{}

	_ InputStorePaymentPurposeClass = &InputStorePaymentPremiumSubscription{}
)

func (i *InputStorePaymentPremiumSubscription) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Flags.Zero()) {
		return false
	}
	if !(i.Restore == false) {
		return false
	}
	if !(i.Upgrade == false) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputStorePaymentPremiumSubscription) String() string {
	if i == nil {
		return "InputStorePaymentPremiumSubscription(nil)"
	}
	type Alias InputStorePaymentPremiumSubscription
	return fmt.Sprintf("InputStorePaymentPremiumSubscription%+v", Alias(*i))
}

// FillFrom fills InputStorePaymentPremiumSubscription from given interface.
func (i *InputStorePaymentPremiumSubscription) FillFrom(from interface {
	GetRestore() (value bool)
	GetUpgrade() (value bool)
}) {
	i.Restore = from.GetRestore()
	i.Upgrade = from.GetUpgrade()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputStorePaymentPremiumSubscription) TypeID() uint32 {
	return InputStorePaymentPremiumSubscriptionTypeID
}

// TypeName returns name of type in TL schema.
func (*InputStorePaymentPremiumSubscription) TypeName() string {
	return "inputStorePaymentPremiumSubscription"
}

// TypeInfo returns info about TL type.
func (i *InputStorePaymentPremiumSubscription) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputStorePaymentPremiumSubscription",
		ID:   InputStorePaymentPremiumSubscriptionTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Restore",
			SchemaName: "restore",
			Null:       !i.Flags.Has(0),
		},
		{
			Name:       "Upgrade",
			SchemaName: "upgrade",
			Null:       !i.Flags.Has(1),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (i *InputStorePaymentPremiumSubscription) SetFlags() {
	if !(i.Restore == false) {
		i.Flags.Set(0)
	}
	if !(i.Upgrade == false) {
		i.Flags.Set(1)
	}
}

// Encode implements bin.Encoder.
func (i *InputStorePaymentPremiumSubscription) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumSubscription#a6751e66 as nil")
	}
	b.PutID(InputStorePaymentPremiumSubscriptionTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputStorePaymentPremiumSubscription) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumSubscription#a6751e66 as nil")
	}
	i.SetFlags()
	if err := i.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputStorePaymentPremiumSubscription#a6751e66: field flags: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (i *InputStorePaymentPremiumSubscription) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumSubscription#a6751e66 to nil")
	}
	if err := b.ConsumeID(InputStorePaymentPremiumSubscriptionTypeID); err != nil {
		return fmt.Errorf("unable to decode inputStorePaymentPremiumSubscription#a6751e66: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputStorePaymentPremiumSubscription) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumSubscription#a6751e66 to nil")
	}
	{
		if err := i.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumSubscription#a6751e66: field flags: %w", err)
		}
	}
	i.Restore = i.Flags.Has(0)
	i.Upgrade = i.Flags.Has(1)
	return nil
}

// SetRestore sets value of Restore conditional field.
func (i *InputStorePaymentPremiumSubscription) SetRestore(value bool) {
	if value {
		i.Flags.Set(0)
		i.Restore = true
	} else {
		i.Flags.Unset(0)
		i.Restore = false
	}
}

// GetRestore returns value of Restore conditional field.
func (i *InputStorePaymentPremiumSubscription) GetRestore() (value bool) {
	if i == nil {
		return
	}
	return i.Flags.Has(0)
}

// SetUpgrade sets value of Upgrade conditional field.
func (i *InputStorePaymentPremiumSubscription) SetUpgrade(value bool) {
	if value {
		i.Flags.Set(1)
		i.Upgrade = true
	} else {
		i.Flags.Unset(1)
		i.Upgrade = false
	}
}

// GetUpgrade returns value of Upgrade conditional field.
func (i *InputStorePaymentPremiumSubscription) GetUpgrade() (value bool) {
	if i == nil {
		return
	}
	return i.Flags.Has(1)
}

// InputStorePaymentGiftPremium represents TL type `inputStorePaymentGiftPremium#616f7fe8`.
// Info about a gifted Telegram Premium purchase
//
// See https://core.telegram.org/constructor/inputStorePaymentGiftPremium for reference.
type InputStorePaymentGiftPremium struct {
	// The user to which the Telegram Premium subscription was gifted
	UserID InputUserClass
	// Three-letter ISO 4217 currency¹ code
	//
	// Links:
	//  1) https://core.telegram.org/bots/payments#supported-currencies
	Currency string
	// Price of the product in the smallest units of the currency (integer, not float/double)
	// For example, for a price of US$ 1.45 pass amount = 145. See the exp parameter in
	// currencies.json¹, it shows the number of digits past the decimal point for each
	// currency (2 for the majority of currencies).
	//
	// Links:
	//  1) https://core.telegram.org/bots/payments/currencies.json
	Amount int64
}

// InputStorePaymentGiftPremiumTypeID is TL type id of InputStorePaymentGiftPremium.
const InputStorePaymentGiftPremiumTypeID = 0x616f7fe8

// construct implements constructor of InputStorePaymentPurposeClass.
func (i InputStorePaymentGiftPremium) construct() InputStorePaymentPurposeClass { return &i }

// Ensuring interfaces in compile-time for InputStorePaymentGiftPremium.
var (
	_ bin.Encoder     = &InputStorePaymentGiftPremium{}
	_ bin.Decoder     = &InputStorePaymentGiftPremium{}
	_ bin.BareEncoder = &InputStorePaymentGiftPremium{}
	_ bin.BareDecoder = &InputStorePaymentGiftPremium{}

	_ InputStorePaymentPurposeClass = &InputStorePaymentGiftPremium{}
)

func (i *InputStorePaymentGiftPremium) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.UserID == nil) {
		return false
	}
	if !(i.Currency == "") {
		return false
	}
	if !(i.Amount == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputStorePaymentGiftPremium) String() string {
	if i == nil {
		return "InputStorePaymentGiftPremium(nil)"
	}
	type Alias InputStorePaymentGiftPremium
	return fmt.Sprintf("InputStorePaymentGiftPremium%+v", Alias(*i))
}

// FillFrom fills InputStorePaymentGiftPremium from given interface.
func (i *InputStorePaymentGiftPremium) FillFrom(from interface {
	GetUserID() (value InputUserClass)
	GetCurrency() (value string)
	GetAmount() (value int64)
}) {
	i.UserID = from.GetUserID()
	i.Currency = from.GetCurrency()
	i.Amount = from.GetAmount()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputStorePaymentGiftPremium) TypeID() uint32 {
	return InputStorePaymentGiftPremiumTypeID
}

// TypeName returns name of type in TL schema.
func (*InputStorePaymentGiftPremium) TypeName() string {
	return "inputStorePaymentGiftPremium"
}

// TypeInfo returns info about TL type.
func (i *InputStorePaymentGiftPremium) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputStorePaymentGiftPremium",
		ID:   InputStorePaymentGiftPremiumTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
		{
			Name:       "Currency",
			SchemaName: "currency",
		},
		{
			Name:       "Amount",
			SchemaName: "amount",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputStorePaymentGiftPremium) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentGiftPremium#616f7fe8 as nil")
	}
	b.PutID(InputStorePaymentGiftPremiumTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputStorePaymentGiftPremium) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentGiftPremium#616f7fe8 as nil")
	}
	if i.UserID == nil {
		return fmt.Errorf("unable to encode inputStorePaymentGiftPremium#616f7fe8: field user_id is nil")
	}
	if err := i.UserID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputStorePaymentGiftPremium#616f7fe8: field user_id: %w", err)
	}
	b.PutString(i.Currency)
	b.PutLong(i.Amount)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputStorePaymentGiftPremium) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentGiftPremium#616f7fe8 to nil")
	}
	if err := b.ConsumeID(InputStorePaymentGiftPremiumTypeID); err != nil {
		return fmt.Errorf("unable to decode inputStorePaymentGiftPremium#616f7fe8: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputStorePaymentGiftPremium) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentGiftPremium#616f7fe8 to nil")
	}
	{
		value, err := DecodeInputUser(b)
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentGiftPremium#616f7fe8: field user_id: %w", err)
		}
		i.UserID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentGiftPremium#616f7fe8: field currency: %w", err)
		}
		i.Currency = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentGiftPremium#616f7fe8: field amount: %w", err)
		}
		i.Amount = value
	}
	return nil
}

// GetUserID returns value of UserID field.
func (i *InputStorePaymentGiftPremium) GetUserID() (value InputUserClass) {
	if i == nil {
		return
	}
	return i.UserID
}

// GetCurrency returns value of Currency field.
func (i *InputStorePaymentGiftPremium) GetCurrency() (value string) {
	if i == nil {
		return
	}
	return i.Currency
}

// GetAmount returns value of Amount field.
func (i *InputStorePaymentGiftPremium) GetAmount() (value int64) {
	if i == nil {
		return
	}
	return i.Amount
}

// InputStorePaymentPremiumGiftCode represents TL type `inputStorePaymentPremiumGiftCode#a3805f3f`.
//
// See https://core.telegram.org/constructor/inputStorePaymentPremiumGiftCode for reference.
type InputStorePaymentPremiumGiftCode struct {
	// Flags field of InputStorePaymentPremiumGiftCode.
	Flags bin.Fields
	// Users field of InputStorePaymentPremiumGiftCode.
	Users []InputUserClass
	// BoostPeer field of InputStorePaymentPremiumGiftCode.
	//
	// Use SetBoostPeer and GetBoostPeer helpers.
	BoostPeer InputPeerClass
	// Currency field of InputStorePaymentPremiumGiftCode.
	Currency string
	// Amount field of InputStorePaymentPremiumGiftCode.
	Amount int64
}

// InputStorePaymentPremiumGiftCodeTypeID is TL type id of InputStorePaymentPremiumGiftCode.
const InputStorePaymentPremiumGiftCodeTypeID = 0xa3805f3f

// construct implements constructor of InputStorePaymentPurposeClass.
func (i InputStorePaymentPremiumGiftCode) construct() InputStorePaymentPurposeClass { return &i }

// Ensuring interfaces in compile-time for InputStorePaymentPremiumGiftCode.
var (
	_ bin.Encoder     = &InputStorePaymentPremiumGiftCode{}
	_ bin.Decoder     = &InputStorePaymentPremiumGiftCode{}
	_ bin.BareEncoder = &InputStorePaymentPremiumGiftCode{}
	_ bin.BareDecoder = &InputStorePaymentPremiumGiftCode{}

	_ InputStorePaymentPurposeClass = &InputStorePaymentPremiumGiftCode{}
)

func (i *InputStorePaymentPremiumGiftCode) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Flags.Zero()) {
		return false
	}
	if !(i.Users == nil) {
		return false
	}
	if !(i.BoostPeer == nil) {
		return false
	}
	if !(i.Currency == "") {
		return false
	}
	if !(i.Amount == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputStorePaymentPremiumGiftCode) String() string {
	if i == nil {
		return "InputStorePaymentPremiumGiftCode(nil)"
	}
	type Alias InputStorePaymentPremiumGiftCode
	return fmt.Sprintf("InputStorePaymentPremiumGiftCode%+v", Alias(*i))
}

// FillFrom fills InputStorePaymentPremiumGiftCode from given interface.
func (i *InputStorePaymentPremiumGiftCode) FillFrom(from interface {
	GetUsers() (value []InputUserClass)
	GetBoostPeer() (value InputPeerClass, ok bool)
	GetCurrency() (value string)
	GetAmount() (value int64)
}) {
	i.Users = from.GetUsers()
	if val, ok := from.GetBoostPeer(); ok {
		i.BoostPeer = val
	}

	i.Currency = from.GetCurrency()
	i.Amount = from.GetAmount()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputStorePaymentPremiumGiftCode) TypeID() uint32 {
	return InputStorePaymentPremiumGiftCodeTypeID
}

// TypeName returns name of type in TL schema.
func (*InputStorePaymentPremiumGiftCode) TypeName() string {
	return "inputStorePaymentPremiumGiftCode"
}

// TypeInfo returns info about TL type.
func (i *InputStorePaymentPremiumGiftCode) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputStorePaymentPremiumGiftCode",
		ID:   InputStorePaymentPremiumGiftCodeTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Users",
			SchemaName: "users",
		},
		{
			Name:       "BoostPeer",
			SchemaName: "boost_peer",
			Null:       !i.Flags.Has(0),
		},
		{
			Name:       "Currency",
			SchemaName: "currency",
		},
		{
			Name:       "Amount",
			SchemaName: "amount",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (i *InputStorePaymentPremiumGiftCode) SetFlags() {
	if !(i.BoostPeer == nil) {
		i.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (i *InputStorePaymentPremiumGiftCode) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumGiftCode#a3805f3f as nil")
	}
	b.PutID(InputStorePaymentPremiumGiftCodeTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputStorePaymentPremiumGiftCode) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumGiftCode#a3805f3f as nil")
	}
	i.SetFlags()
	if err := i.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputStorePaymentPremiumGiftCode#a3805f3f: field flags: %w", err)
	}
	b.PutVectorHeader(len(i.Users))
	for idx, v := range i.Users {
		if v == nil {
			return fmt.Errorf("unable to encode inputStorePaymentPremiumGiftCode#a3805f3f: field users element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode inputStorePaymentPremiumGiftCode#a3805f3f: field users element with index %d: %w", idx, err)
		}
	}
	if i.Flags.Has(0) {
		if i.BoostPeer == nil {
			return fmt.Errorf("unable to encode inputStorePaymentPremiumGiftCode#a3805f3f: field boost_peer is nil")
		}
		if err := i.BoostPeer.Encode(b); err != nil {
			return fmt.Errorf("unable to encode inputStorePaymentPremiumGiftCode#a3805f3f: field boost_peer: %w", err)
		}
	}
	b.PutString(i.Currency)
	b.PutLong(i.Amount)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputStorePaymentPremiumGiftCode) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumGiftCode#a3805f3f to nil")
	}
	if err := b.ConsumeID(InputStorePaymentPremiumGiftCodeTypeID); err != nil {
		return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputStorePaymentPremiumGiftCode) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumGiftCode#a3805f3f to nil")
	}
	{
		if err := i.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field flags: %w", err)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field users: %w", err)
		}

		if headerLen > 0 {
			i.Users = make([]InputUserClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputUser(b)
			if err != nil {
				return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field users: %w", err)
			}
			i.Users = append(i.Users, value)
		}
	}
	if i.Flags.Has(0) {
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field boost_peer: %w", err)
		}
		i.BoostPeer = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field currency: %w", err)
		}
		i.Currency = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiftCode#a3805f3f: field amount: %w", err)
		}
		i.Amount = value
	}
	return nil
}

// GetUsers returns value of Users field.
func (i *InputStorePaymentPremiumGiftCode) GetUsers() (value []InputUserClass) {
	if i == nil {
		return
	}
	return i.Users
}

// SetBoostPeer sets value of BoostPeer conditional field.
func (i *InputStorePaymentPremiumGiftCode) SetBoostPeer(value InputPeerClass) {
	i.Flags.Set(0)
	i.BoostPeer = value
}

// GetBoostPeer returns value of BoostPeer conditional field and
// boolean which is true if field was set.
func (i *InputStorePaymentPremiumGiftCode) GetBoostPeer() (value InputPeerClass, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(0) {
		return value, false
	}
	return i.BoostPeer, true
}

// GetCurrency returns value of Currency field.
func (i *InputStorePaymentPremiumGiftCode) GetCurrency() (value string) {
	if i == nil {
		return
	}
	return i.Currency
}

// GetAmount returns value of Amount field.
func (i *InputStorePaymentPremiumGiftCode) GetAmount() (value int64) {
	if i == nil {
		return
	}
	return i.Amount
}

// MapUsers returns field Users wrapped in InputUserClassArray helper.
func (i *InputStorePaymentPremiumGiftCode) MapUsers() (value InputUserClassArray) {
	return InputUserClassArray(i.Users)
}

// InputStorePaymentPremiumGiveaway represents TL type `inputStorePaymentPremiumGiveaway#7c9375e6`.
//
// See https://core.telegram.org/constructor/inputStorePaymentPremiumGiveaway for reference.
type InputStorePaymentPremiumGiveaway struct {
	// Flags field of InputStorePaymentPremiumGiveaway.
	Flags bin.Fields
	// OnlyNewSubscribers field of InputStorePaymentPremiumGiveaway.
	OnlyNewSubscribers bool
	// BoostPeer field of InputStorePaymentPremiumGiveaway.
	BoostPeer InputPeerClass
	// AdditionalPeers field of InputStorePaymentPremiumGiveaway.
	//
	// Use SetAdditionalPeers and GetAdditionalPeers helpers.
	AdditionalPeers []InputPeerClass
	// CountriesISO2 field of InputStorePaymentPremiumGiveaway.
	//
	// Use SetCountriesISO2 and GetCountriesISO2 helpers.
	CountriesISO2 []string
	// RandomID field of InputStorePaymentPremiumGiveaway.
	RandomID int64
	// UntilDate field of InputStorePaymentPremiumGiveaway.
	UntilDate int
	// Currency field of InputStorePaymentPremiumGiveaway.
	Currency string
	// Amount field of InputStorePaymentPremiumGiveaway.
	Amount int64
}

// InputStorePaymentPremiumGiveawayTypeID is TL type id of InputStorePaymentPremiumGiveaway.
const InputStorePaymentPremiumGiveawayTypeID = 0x7c9375e6

// construct implements constructor of InputStorePaymentPurposeClass.
func (i InputStorePaymentPremiumGiveaway) construct() InputStorePaymentPurposeClass { return &i }

// Ensuring interfaces in compile-time for InputStorePaymentPremiumGiveaway.
var (
	_ bin.Encoder     = &InputStorePaymentPremiumGiveaway{}
	_ bin.Decoder     = &InputStorePaymentPremiumGiveaway{}
	_ bin.BareEncoder = &InputStorePaymentPremiumGiveaway{}
	_ bin.BareDecoder = &InputStorePaymentPremiumGiveaway{}

	_ InputStorePaymentPurposeClass = &InputStorePaymentPremiumGiveaway{}
)

func (i *InputStorePaymentPremiumGiveaway) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Flags.Zero()) {
		return false
	}
	if !(i.OnlyNewSubscribers == false) {
		return false
	}
	if !(i.BoostPeer == nil) {
		return false
	}
	if !(i.AdditionalPeers == nil) {
		return false
	}
	if !(i.CountriesISO2 == nil) {
		return false
	}
	if !(i.RandomID == 0) {
		return false
	}
	if !(i.UntilDate == 0) {
		return false
	}
	if !(i.Currency == "") {
		return false
	}
	if !(i.Amount == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputStorePaymentPremiumGiveaway) String() string {
	if i == nil {
		return "InputStorePaymentPremiumGiveaway(nil)"
	}
	type Alias InputStorePaymentPremiumGiveaway
	return fmt.Sprintf("InputStorePaymentPremiumGiveaway%+v", Alias(*i))
}

// FillFrom fills InputStorePaymentPremiumGiveaway from given interface.
func (i *InputStorePaymentPremiumGiveaway) FillFrom(from interface {
	GetOnlyNewSubscribers() (value bool)
	GetBoostPeer() (value InputPeerClass)
	GetAdditionalPeers() (value []InputPeerClass, ok bool)
	GetCountriesISO2() (value []string, ok bool)
	GetRandomID() (value int64)
	GetUntilDate() (value int)
	GetCurrency() (value string)
	GetAmount() (value int64)
}) {
	i.OnlyNewSubscribers = from.GetOnlyNewSubscribers()
	i.BoostPeer = from.GetBoostPeer()
	if val, ok := from.GetAdditionalPeers(); ok {
		i.AdditionalPeers = val
	}

	if val, ok := from.GetCountriesISO2(); ok {
		i.CountriesISO2 = val
	}

	i.RandomID = from.GetRandomID()
	i.UntilDate = from.GetUntilDate()
	i.Currency = from.GetCurrency()
	i.Amount = from.GetAmount()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputStorePaymentPremiumGiveaway) TypeID() uint32 {
	return InputStorePaymentPremiumGiveawayTypeID
}

// TypeName returns name of type in TL schema.
func (*InputStorePaymentPremiumGiveaway) TypeName() string {
	return "inputStorePaymentPremiumGiveaway"
}

// TypeInfo returns info about TL type.
func (i *InputStorePaymentPremiumGiveaway) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputStorePaymentPremiumGiveaway",
		ID:   InputStorePaymentPremiumGiveawayTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "OnlyNewSubscribers",
			SchemaName: "only_new_subscribers",
			Null:       !i.Flags.Has(0),
		},
		{
			Name:       "BoostPeer",
			SchemaName: "boost_peer",
		},
		{
			Name:       "AdditionalPeers",
			SchemaName: "additional_peers",
			Null:       !i.Flags.Has(1),
		},
		{
			Name:       "CountriesISO2",
			SchemaName: "countries_iso2",
			Null:       !i.Flags.Has(2),
		},
		{
			Name:       "RandomID",
			SchemaName: "random_id",
		},
		{
			Name:       "UntilDate",
			SchemaName: "until_date",
		},
		{
			Name:       "Currency",
			SchemaName: "currency",
		},
		{
			Name:       "Amount",
			SchemaName: "amount",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (i *InputStorePaymentPremiumGiveaway) SetFlags() {
	if !(i.OnlyNewSubscribers == false) {
		i.Flags.Set(0)
	}
	if !(i.AdditionalPeers == nil) {
		i.Flags.Set(1)
	}
	if !(i.CountriesISO2 == nil) {
		i.Flags.Set(2)
	}
}

// Encode implements bin.Encoder.
func (i *InputStorePaymentPremiumGiveaway) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumGiveaway#7c9375e6 as nil")
	}
	b.PutID(InputStorePaymentPremiumGiveawayTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputStorePaymentPremiumGiveaway) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputStorePaymentPremiumGiveaway#7c9375e6 as nil")
	}
	i.SetFlags()
	if err := i.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputStorePaymentPremiumGiveaway#7c9375e6: field flags: %w", err)
	}
	if i.BoostPeer == nil {
		return fmt.Errorf("unable to encode inputStorePaymentPremiumGiveaway#7c9375e6: field boost_peer is nil")
	}
	if err := i.BoostPeer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputStorePaymentPremiumGiveaway#7c9375e6: field boost_peer: %w", err)
	}
	if i.Flags.Has(1) {
		b.PutVectorHeader(len(i.AdditionalPeers))
		for idx, v := range i.AdditionalPeers {
			if v == nil {
				return fmt.Errorf("unable to encode inputStorePaymentPremiumGiveaway#7c9375e6: field additional_peers element with index %d is nil", idx)
			}
			if err := v.Encode(b); err != nil {
				return fmt.Errorf("unable to encode inputStorePaymentPremiumGiveaway#7c9375e6: field additional_peers element with index %d: %w", idx, err)
			}
		}
	}
	if i.Flags.Has(2) {
		b.PutVectorHeader(len(i.CountriesISO2))
		for _, v := range i.CountriesISO2 {
			b.PutString(v)
		}
	}
	b.PutLong(i.RandomID)
	b.PutInt(i.UntilDate)
	b.PutString(i.Currency)
	b.PutLong(i.Amount)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputStorePaymentPremiumGiveaway) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumGiveaway#7c9375e6 to nil")
	}
	if err := b.ConsumeID(InputStorePaymentPremiumGiveawayTypeID); err != nil {
		return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputStorePaymentPremiumGiveaway) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputStorePaymentPremiumGiveaway#7c9375e6 to nil")
	}
	{
		if err := i.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field flags: %w", err)
		}
	}
	i.OnlyNewSubscribers = i.Flags.Has(0)
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field boost_peer: %w", err)
		}
		i.BoostPeer = value
	}
	if i.Flags.Has(1) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field additional_peers: %w", err)
		}

		if headerLen > 0 {
			i.AdditionalPeers = make([]InputPeerClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputPeer(b)
			if err != nil {
				return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field additional_peers: %w", err)
			}
			i.AdditionalPeers = append(i.AdditionalPeers, value)
		}
	}
	if i.Flags.Has(2) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field countries_iso2: %w", err)
		}

		if headerLen > 0 {
			i.CountriesISO2 = make([]string, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field countries_iso2: %w", err)
			}
			i.CountriesISO2 = append(i.CountriesISO2, value)
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field random_id: %w", err)
		}
		i.RandomID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field until_date: %w", err)
		}
		i.UntilDate = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field currency: %w", err)
		}
		i.Currency = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode inputStorePaymentPremiumGiveaway#7c9375e6: field amount: %w", err)
		}
		i.Amount = value
	}
	return nil
}

// SetOnlyNewSubscribers sets value of OnlyNewSubscribers conditional field.
func (i *InputStorePaymentPremiumGiveaway) SetOnlyNewSubscribers(value bool) {
	if value {
		i.Flags.Set(0)
		i.OnlyNewSubscribers = true
	} else {
		i.Flags.Unset(0)
		i.OnlyNewSubscribers = false
	}
}

// GetOnlyNewSubscribers returns value of OnlyNewSubscribers conditional field.
func (i *InputStorePaymentPremiumGiveaway) GetOnlyNewSubscribers() (value bool) {
	if i == nil {
		return
	}
	return i.Flags.Has(0)
}

// GetBoostPeer returns value of BoostPeer field.
func (i *InputStorePaymentPremiumGiveaway) GetBoostPeer() (value InputPeerClass) {
	if i == nil {
		return
	}
	return i.BoostPeer
}

// SetAdditionalPeers sets value of AdditionalPeers conditional field.
func (i *InputStorePaymentPremiumGiveaway) SetAdditionalPeers(value []InputPeerClass) {
	i.Flags.Set(1)
	i.AdditionalPeers = value
}

// GetAdditionalPeers returns value of AdditionalPeers conditional field and
// boolean which is true if field was set.
func (i *InputStorePaymentPremiumGiveaway) GetAdditionalPeers() (value []InputPeerClass, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(1) {
		return value, false
	}
	return i.AdditionalPeers, true
}

// SetCountriesISO2 sets value of CountriesISO2 conditional field.
func (i *InputStorePaymentPremiumGiveaway) SetCountriesISO2(value []string) {
	i.Flags.Set(2)
	i.CountriesISO2 = value
}

// GetCountriesISO2 returns value of CountriesISO2 conditional field and
// boolean which is true if field was set.
func (i *InputStorePaymentPremiumGiveaway) GetCountriesISO2() (value []string, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(2) {
		return value, false
	}
	return i.CountriesISO2, true
}

// GetRandomID returns value of RandomID field.
func (i *InputStorePaymentPremiumGiveaway) GetRandomID() (value int64) {
	if i == nil {
		return
	}
	return i.RandomID
}

// GetUntilDate returns value of UntilDate field.
func (i *InputStorePaymentPremiumGiveaway) GetUntilDate() (value int) {
	if i == nil {
		return
	}
	return i.UntilDate
}

// GetCurrency returns value of Currency field.
func (i *InputStorePaymentPremiumGiveaway) GetCurrency() (value string) {
	if i == nil {
		return
	}
	return i.Currency
}

// GetAmount returns value of Amount field.
func (i *InputStorePaymentPremiumGiveaway) GetAmount() (value int64) {
	if i == nil {
		return
	}
	return i.Amount
}

// MapAdditionalPeers returns field AdditionalPeers wrapped in InputPeerClassArray helper.
func (i *InputStorePaymentPremiumGiveaway) MapAdditionalPeers() (value InputPeerClassArray, ok bool) {
	if !i.Flags.Has(1) {
		return value, false
	}
	return InputPeerClassArray(i.AdditionalPeers), true
}

// InputStorePaymentPurposeClassName is schema name of InputStorePaymentPurposeClass.
const InputStorePaymentPurposeClassName = "InputStorePaymentPurpose"

// InputStorePaymentPurposeClass represents InputStorePaymentPurpose generic type.
//
// See https://core.telegram.org/type/InputStorePaymentPurpose for reference.
//
// Example:
//
//	g, err := tg.DecodeInputStorePaymentPurpose(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tg.InputStorePaymentPremiumSubscription: // inputStorePaymentPremiumSubscription#a6751e66
//	case *tg.InputStorePaymentGiftPremium: // inputStorePaymentGiftPremium#616f7fe8
//	case *tg.InputStorePaymentPremiumGiftCode: // inputStorePaymentPremiumGiftCode#a3805f3f
//	case *tg.InputStorePaymentPremiumGiveaway: // inputStorePaymentPremiumGiveaway#7c9375e6
//	default: panic(v)
//	}
type InputStorePaymentPurposeClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() InputStorePaymentPurposeClass

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

// DecodeInputStorePaymentPurpose implements binary de-serialization for InputStorePaymentPurposeClass.
func DecodeInputStorePaymentPurpose(buf *bin.Buffer) (InputStorePaymentPurposeClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case InputStorePaymentPremiumSubscriptionTypeID:
		// Decoding inputStorePaymentPremiumSubscription#a6751e66.
		v := InputStorePaymentPremiumSubscription{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputStorePaymentPurposeClass: %w", err)
		}
		return &v, nil
	case InputStorePaymentGiftPremiumTypeID:
		// Decoding inputStorePaymentGiftPremium#616f7fe8.
		v := InputStorePaymentGiftPremium{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputStorePaymentPurposeClass: %w", err)
		}
		return &v, nil
	case InputStorePaymentPremiumGiftCodeTypeID:
		// Decoding inputStorePaymentPremiumGiftCode#a3805f3f.
		v := InputStorePaymentPremiumGiftCode{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputStorePaymentPurposeClass: %w", err)
		}
		return &v, nil
	case InputStorePaymentPremiumGiveawayTypeID:
		// Decoding inputStorePaymentPremiumGiveaway#7c9375e6.
		v := InputStorePaymentPremiumGiveaway{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode InputStorePaymentPurposeClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode InputStorePaymentPurposeClass: %w", bin.NewUnexpectedID(id))
	}
}

// InputStorePaymentPurpose boxes the InputStorePaymentPurposeClass providing a helper.
type InputStorePaymentPurposeBox struct {
	InputStorePaymentPurpose InputStorePaymentPurposeClass
}

// Decode implements bin.Decoder for InputStorePaymentPurposeBox.
func (b *InputStorePaymentPurposeBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode InputStorePaymentPurposeBox to nil")
	}
	v, err := DecodeInputStorePaymentPurpose(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.InputStorePaymentPurpose = v
	return nil
}

// Encode implements bin.Encode for InputStorePaymentPurposeBox.
func (b *InputStorePaymentPurposeBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.InputStorePaymentPurpose == nil {
		return fmt.Errorf("unable to encode InputStorePaymentPurposeClass as nil")
	}
	return b.InputStorePaymentPurpose.Encode(buf)
}
