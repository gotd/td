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

// PaymentsGetSavedStarGiftsRequest represents TL type `payments.getSavedStarGifts#23830de9`.
//
// See https://core.telegram.org/method/payments.getSavedStarGifts for reference.
type PaymentsGetSavedStarGiftsRequest struct {
	// Flags field of PaymentsGetSavedStarGiftsRequest.
	Flags bin.Fields
	// ExcludeUnsaved field of PaymentsGetSavedStarGiftsRequest.
	ExcludeUnsaved bool
	// ExcludeSaved field of PaymentsGetSavedStarGiftsRequest.
	ExcludeSaved bool
	// ExcludeUnlimited field of PaymentsGetSavedStarGiftsRequest.
	ExcludeUnlimited bool
	// ExcludeLimited field of PaymentsGetSavedStarGiftsRequest.
	ExcludeLimited bool
	// ExcludeUnique field of PaymentsGetSavedStarGiftsRequest.
	ExcludeUnique bool
	// SortByValue field of PaymentsGetSavedStarGiftsRequest.
	SortByValue bool
	// Peer field of PaymentsGetSavedStarGiftsRequest.
	Peer InputPeerClass
	// Offset field of PaymentsGetSavedStarGiftsRequest.
	Offset string
	// Limit field of PaymentsGetSavedStarGiftsRequest.
	Limit int
}

// PaymentsGetSavedStarGiftsRequestTypeID is TL type id of PaymentsGetSavedStarGiftsRequest.
const PaymentsGetSavedStarGiftsRequestTypeID = 0x23830de9

// Ensuring interfaces in compile-time for PaymentsGetSavedStarGiftsRequest.
var (
	_ bin.Encoder     = &PaymentsGetSavedStarGiftsRequest{}
	_ bin.Decoder     = &PaymentsGetSavedStarGiftsRequest{}
	_ bin.BareEncoder = &PaymentsGetSavedStarGiftsRequest{}
	_ bin.BareDecoder = &PaymentsGetSavedStarGiftsRequest{}
)

func (g *PaymentsGetSavedStarGiftsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.ExcludeUnsaved == false) {
		return false
	}
	if !(g.ExcludeSaved == false) {
		return false
	}
	if !(g.ExcludeUnlimited == false) {
		return false
	}
	if !(g.ExcludeLimited == false) {
		return false
	}
	if !(g.ExcludeUnique == false) {
		return false
	}
	if !(g.SortByValue == false) {
		return false
	}
	if !(g.Peer == nil) {
		return false
	}
	if !(g.Offset == "") {
		return false
	}
	if !(g.Limit == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *PaymentsGetSavedStarGiftsRequest) String() string {
	if g == nil {
		return "PaymentsGetSavedStarGiftsRequest(nil)"
	}
	type Alias PaymentsGetSavedStarGiftsRequest
	return fmt.Sprintf("PaymentsGetSavedStarGiftsRequest%+v", Alias(*g))
}

// FillFrom fills PaymentsGetSavedStarGiftsRequest from given interface.
func (g *PaymentsGetSavedStarGiftsRequest) FillFrom(from interface {
	GetExcludeUnsaved() (value bool)
	GetExcludeSaved() (value bool)
	GetExcludeUnlimited() (value bool)
	GetExcludeLimited() (value bool)
	GetExcludeUnique() (value bool)
	GetSortByValue() (value bool)
	GetPeer() (value InputPeerClass)
	GetOffset() (value string)
	GetLimit() (value int)
}) {
	g.ExcludeUnsaved = from.GetExcludeUnsaved()
	g.ExcludeSaved = from.GetExcludeSaved()
	g.ExcludeUnlimited = from.GetExcludeUnlimited()
	g.ExcludeLimited = from.GetExcludeLimited()
	g.ExcludeUnique = from.GetExcludeUnique()
	g.SortByValue = from.GetSortByValue()
	g.Peer = from.GetPeer()
	g.Offset = from.GetOffset()
	g.Limit = from.GetLimit()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PaymentsGetSavedStarGiftsRequest) TypeID() uint32 {
	return PaymentsGetSavedStarGiftsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PaymentsGetSavedStarGiftsRequest) TypeName() string {
	return "payments.getSavedStarGifts"
}

// TypeInfo returns info about TL type.
func (g *PaymentsGetSavedStarGiftsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "payments.getSavedStarGifts",
		ID:   PaymentsGetSavedStarGiftsRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ExcludeUnsaved",
			SchemaName: "exclude_unsaved",
			Null:       !g.Flags.Has(0),
		},
		{
			Name:       "ExcludeSaved",
			SchemaName: "exclude_saved",
			Null:       !g.Flags.Has(1),
		},
		{
			Name:       "ExcludeUnlimited",
			SchemaName: "exclude_unlimited",
			Null:       !g.Flags.Has(2),
		},
		{
			Name:       "ExcludeLimited",
			SchemaName: "exclude_limited",
			Null:       !g.Flags.Has(3),
		},
		{
			Name:       "ExcludeUnique",
			SchemaName: "exclude_unique",
			Null:       !g.Flags.Has(4),
		},
		{
			Name:       "SortByValue",
			SchemaName: "sort_by_value",
			Null:       !g.Flags.Has(5),
		},
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "Offset",
			SchemaName: "offset",
		},
		{
			Name:       "Limit",
			SchemaName: "limit",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (g *PaymentsGetSavedStarGiftsRequest) SetFlags() {
	if !(g.ExcludeUnsaved == false) {
		g.Flags.Set(0)
	}
	if !(g.ExcludeSaved == false) {
		g.Flags.Set(1)
	}
	if !(g.ExcludeUnlimited == false) {
		g.Flags.Set(2)
	}
	if !(g.ExcludeLimited == false) {
		g.Flags.Set(3)
	}
	if !(g.ExcludeUnique == false) {
		g.Flags.Set(4)
	}
	if !(g.SortByValue == false) {
		g.Flags.Set(5)
	}
}

// Encode implements bin.Encoder.
func (g *PaymentsGetSavedStarGiftsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getSavedStarGifts#23830de9 as nil")
	}
	b.PutID(PaymentsGetSavedStarGiftsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PaymentsGetSavedStarGiftsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getSavedStarGifts#23830de9 as nil")
	}
	g.SetFlags()
	if err := g.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getSavedStarGifts#23830de9: field flags: %w", err)
	}
	if g.Peer == nil {
		return fmt.Errorf("unable to encode payments.getSavedStarGifts#23830de9: field peer is nil")
	}
	if err := g.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getSavedStarGifts#23830de9: field peer: %w", err)
	}
	b.PutString(g.Offset)
	b.PutInt(g.Limit)
	return nil
}

// Decode implements bin.Decoder.
func (g *PaymentsGetSavedStarGiftsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getSavedStarGifts#23830de9 to nil")
	}
	if err := b.ConsumeID(PaymentsGetSavedStarGiftsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode payments.getSavedStarGifts#23830de9: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PaymentsGetSavedStarGiftsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getSavedStarGifts#23830de9 to nil")
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode payments.getSavedStarGifts#23830de9: field flags: %w", err)
		}
	}
	g.ExcludeUnsaved = g.Flags.Has(0)
	g.ExcludeSaved = g.Flags.Has(1)
	g.ExcludeUnlimited = g.Flags.Has(2)
	g.ExcludeLimited = g.Flags.Has(3)
	g.ExcludeUnique = g.Flags.Has(4)
	g.SortByValue = g.Flags.Has(5)
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode payments.getSavedStarGifts#23830de9: field peer: %w", err)
		}
		g.Peer = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode payments.getSavedStarGifts#23830de9: field offset: %w", err)
		}
		g.Offset = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode payments.getSavedStarGifts#23830de9: field limit: %w", err)
		}
		g.Limit = value
	}
	return nil
}

// SetExcludeUnsaved sets value of ExcludeUnsaved conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetExcludeUnsaved(value bool) {
	if value {
		g.Flags.Set(0)
		g.ExcludeUnsaved = true
	} else {
		g.Flags.Unset(0)
		g.ExcludeUnsaved = false
	}
}

// GetExcludeUnsaved returns value of ExcludeUnsaved conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetExcludeUnsaved() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(0)
}

// SetExcludeSaved sets value of ExcludeSaved conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetExcludeSaved(value bool) {
	if value {
		g.Flags.Set(1)
		g.ExcludeSaved = true
	} else {
		g.Flags.Unset(1)
		g.ExcludeSaved = false
	}
}

// GetExcludeSaved returns value of ExcludeSaved conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetExcludeSaved() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(1)
}

// SetExcludeUnlimited sets value of ExcludeUnlimited conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetExcludeUnlimited(value bool) {
	if value {
		g.Flags.Set(2)
		g.ExcludeUnlimited = true
	} else {
		g.Flags.Unset(2)
		g.ExcludeUnlimited = false
	}
}

// GetExcludeUnlimited returns value of ExcludeUnlimited conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetExcludeUnlimited() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(2)
}

// SetExcludeLimited sets value of ExcludeLimited conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetExcludeLimited(value bool) {
	if value {
		g.Flags.Set(3)
		g.ExcludeLimited = true
	} else {
		g.Flags.Unset(3)
		g.ExcludeLimited = false
	}
}

// GetExcludeLimited returns value of ExcludeLimited conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetExcludeLimited() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(3)
}

// SetExcludeUnique sets value of ExcludeUnique conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetExcludeUnique(value bool) {
	if value {
		g.Flags.Set(4)
		g.ExcludeUnique = true
	} else {
		g.Flags.Unset(4)
		g.ExcludeUnique = false
	}
}

// GetExcludeUnique returns value of ExcludeUnique conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetExcludeUnique() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(4)
}

// SetSortByValue sets value of SortByValue conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) SetSortByValue(value bool) {
	if value {
		g.Flags.Set(5)
		g.SortByValue = true
	} else {
		g.Flags.Unset(5)
		g.SortByValue = false
	}
}

// GetSortByValue returns value of SortByValue conditional field.
func (g *PaymentsGetSavedStarGiftsRequest) GetSortByValue() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(5)
}

// GetPeer returns value of Peer field.
func (g *PaymentsGetSavedStarGiftsRequest) GetPeer() (value InputPeerClass) {
	if g == nil {
		return
	}
	return g.Peer
}

// GetOffset returns value of Offset field.
func (g *PaymentsGetSavedStarGiftsRequest) GetOffset() (value string) {
	if g == nil {
		return
	}
	return g.Offset
}

// GetLimit returns value of Limit field.
func (g *PaymentsGetSavedStarGiftsRequest) GetLimit() (value int) {
	if g == nil {
		return
	}
	return g.Limit
}

// PaymentsGetSavedStarGifts invokes method payments.getSavedStarGifts#23830de9 returning error if any.
//
// See https://core.telegram.org/method/payments.getSavedStarGifts for reference.
func (c *Client) PaymentsGetSavedStarGifts(ctx context.Context, request *PaymentsGetSavedStarGiftsRequest) (*PaymentsSavedStarGifts, error) {
	var result PaymentsSavedStarGifts

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}