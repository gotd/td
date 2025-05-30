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

// PaymentsGetConnectedStarRefBotsRequest represents TL type `payments.getConnectedStarRefBots#5869a553`.
// Fetch all affiliations we have created for a certain peer
//
// See https://core.telegram.org/method/payments.getConnectedStarRefBots for reference.
type PaymentsGetConnectedStarRefBotsRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// The affiliated peer
	Peer InputPeerClass
	// If set, returns only results older than the specified unixtime
	//
	// Use SetOffsetDate and GetOffsetDate helpers.
	OffsetDate int
	// Offset for pagination¹, taken from the last returned connectedBotStarRef².url
	// (initially empty)
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets
	//  2) https://core.telegram.org/constructor/connectedBotStarRef
	//
	// Use SetOffsetLink and GetOffsetLink helpers.
	OffsetLink string
	// Maximum number of results to return, see pagination¹
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets
	Limit int
}

// PaymentsGetConnectedStarRefBotsRequestTypeID is TL type id of PaymentsGetConnectedStarRefBotsRequest.
const PaymentsGetConnectedStarRefBotsRequestTypeID = 0x5869a553

// Ensuring interfaces in compile-time for PaymentsGetConnectedStarRefBotsRequest.
var (
	_ bin.Encoder     = &PaymentsGetConnectedStarRefBotsRequest{}
	_ bin.Decoder     = &PaymentsGetConnectedStarRefBotsRequest{}
	_ bin.BareEncoder = &PaymentsGetConnectedStarRefBotsRequest{}
	_ bin.BareDecoder = &PaymentsGetConnectedStarRefBotsRequest{}
)

func (g *PaymentsGetConnectedStarRefBotsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.Peer == nil) {
		return false
	}
	if !(g.OffsetDate == 0) {
		return false
	}
	if !(g.OffsetLink == "") {
		return false
	}
	if !(g.Limit == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *PaymentsGetConnectedStarRefBotsRequest) String() string {
	if g == nil {
		return "PaymentsGetConnectedStarRefBotsRequest(nil)"
	}
	type Alias PaymentsGetConnectedStarRefBotsRequest
	return fmt.Sprintf("PaymentsGetConnectedStarRefBotsRequest%+v", Alias(*g))
}

// FillFrom fills PaymentsGetConnectedStarRefBotsRequest from given interface.
func (g *PaymentsGetConnectedStarRefBotsRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
	GetOffsetDate() (value int, ok bool)
	GetOffsetLink() (value string, ok bool)
	GetLimit() (value int)
}) {
	g.Peer = from.GetPeer()
	if val, ok := from.GetOffsetDate(); ok {
		g.OffsetDate = val
	}

	if val, ok := from.GetOffsetLink(); ok {
		g.OffsetLink = val
	}

	g.Limit = from.GetLimit()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PaymentsGetConnectedStarRefBotsRequest) TypeID() uint32 {
	return PaymentsGetConnectedStarRefBotsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PaymentsGetConnectedStarRefBotsRequest) TypeName() string {
	return "payments.getConnectedStarRefBots"
}

// TypeInfo returns info about TL type.
func (g *PaymentsGetConnectedStarRefBotsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "payments.getConnectedStarRefBots",
		ID:   PaymentsGetConnectedStarRefBotsRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "OffsetDate",
			SchemaName: "offset_date",
			Null:       !g.Flags.Has(2),
		},
		{
			Name:       "OffsetLink",
			SchemaName: "offset_link",
			Null:       !g.Flags.Has(2),
		},
		{
			Name:       "Limit",
			SchemaName: "limit",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (g *PaymentsGetConnectedStarRefBotsRequest) SetFlags() {
	if !(g.OffsetDate == 0) {
		g.Flags.Set(2)
	}
	if !(g.OffsetLink == "") {
		g.Flags.Set(2)
	}
}

// Encode implements bin.Encoder.
func (g *PaymentsGetConnectedStarRefBotsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getConnectedStarRefBots#5869a553 as nil")
	}
	b.PutID(PaymentsGetConnectedStarRefBotsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PaymentsGetConnectedStarRefBotsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getConnectedStarRefBots#5869a553 as nil")
	}
	g.SetFlags()
	if err := g.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getConnectedStarRefBots#5869a553: field flags: %w", err)
	}
	if g.Peer == nil {
		return fmt.Errorf("unable to encode payments.getConnectedStarRefBots#5869a553: field peer is nil")
	}
	if err := g.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getConnectedStarRefBots#5869a553: field peer: %w", err)
	}
	if g.Flags.Has(2) {
		b.PutInt(g.OffsetDate)
	}
	if g.Flags.Has(2) {
		b.PutString(g.OffsetLink)
	}
	b.PutInt(g.Limit)
	return nil
}

// Decode implements bin.Decoder.
func (g *PaymentsGetConnectedStarRefBotsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getConnectedStarRefBots#5869a553 to nil")
	}
	if err := b.ConsumeID(PaymentsGetConnectedStarRefBotsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PaymentsGetConnectedStarRefBotsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getConnectedStarRefBots#5869a553 to nil")
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: field flags: %w", err)
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: field peer: %w", err)
		}
		g.Peer = value
	}
	if g.Flags.Has(2) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: field offset_date: %w", err)
		}
		g.OffsetDate = value
	}
	if g.Flags.Has(2) {
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: field offset_link: %w", err)
		}
		g.OffsetLink = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode payments.getConnectedStarRefBots#5869a553: field limit: %w", err)
		}
		g.Limit = value
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *PaymentsGetConnectedStarRefBotsRequest) GetPeer() (value InputPeerClass) {
	if g == nil {
		return
	}
	return g.Peer
}

// SetOffsetDate sets value of OffsetDate conditional field.
func (g *PaymentsGetConnectedStarRefBotsRequest) SetOffsetDate(value int) {
	g.Flags.Set(2)
	g.OffsetDate = value
}

// GetOffsetDate returns value of OffsetDate conditional field and
// boolean which is true if field was set.
func (g *PaymentsGetConnectedStarRefBotsRequest) GetOffsetDate() (value int, ok bool) {
	if g == nil {
		return
	}
	if !g.Flags.Has(2) {
		return value, false
	}
	return g.OffsetDate, true
}

// SetOffsetLink sets value of OffsetLink conditional field.
func (g *PaymentsGetConnectedStarRefBotsRequest) SetOffsetLink(value string) {
	g.Flags.Set(2)
	g.OffsetLink = value
}

// GetOffsetLink returns value of OffsetLink conditional field and
// boolean which is true if field was set.
func (g *PaymentsGetConnectedStarRefBotsRequest) GetOffsetLink() (value string, ok bool) {
	if g == nil {
		return
	}
	if !g.Flags.Has(2) {
		return value, false
	}
	return g.OffsetLink, true
}

// GetLimit returns value of Limit field.
func (g *PaymentsGetConnectedStarRefBotsRequest) GetLimit() (value int) {
	if g == nil {
		return
	}
	return g.Limit
}

// PaymentsGetConnectedStarRefBots invokes method payments.getConnectedStarRefBots#5869a553 returning error if any.
// Fetch all affiliations we have created for a certain peer
//
// See https://core.telegram.org/method/payments.getConnectedStarRefBots for reference.
func (c *Client) PaymentsGetConnectedStarRefBots(ctx context.Context, request *PaymentsGetConnectedStarRefBotsRequest) (*PaymentsConnectedStarRefBots, error) {
	var result PaymentsConnectedStarRefBots

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
