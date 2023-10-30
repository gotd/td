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

// PremiumGetBoostsStatusRequest represents TL type `premium.getBoostsStatus#42f1f61`.
//
// See https://core.telegram.org/method/premium.getBoostsStatus for reference.
type PremiumGetBoostsStatusRequest struct {
	// Peer field of PremiumGetBoostsStatusRequest.
	Peer InputPeerClass
}

// PremiumGetBoostsStatusRequestTypeID is TL type id of PremiumGetBoostsStatusRequest.
const PremiumGetBoostsStatusRequestTypeID = 0x42f1f61

// Ensuring interfaces in compile-time for PremiumGetBoostsStatusRequest.
var (
	_ bin.Encoder     = &PremiumGetBoostsStatusRequest{}
	_ bin.Decoder     = &PremiumGetBoostsStatusRequest{}
	_ bin.BareEncoder = &PremiumGetBoostsStatusRequest{}
	_ bin.BareDecoder = &PremiumGetBoostsStatusRequest{}
)

func (g *PremiumGetBoostsStatusRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Peer == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *PremiumGetBoostsStatusRequest) String() string {
	if g == nil {
		return "PremiumGetBoostsStatusRequest(nil)"
	}
	type Alias PremiumGetBoostsStatusRequest
	return fmt.Sprintf("PremiumGetBoostsStatusRequest%+v", Alias(*g))
}

// FillFrom fills PremiumGetBoostsStatusRequest from given interface.
func (g *PremiumGetBoostsStatusRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
}) {
	g.Peer = from.GetPeer()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PremiumGetBoostsStatusRequest) TypeID() uint32 {
	return PremiumGetBoostsStatusRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PremiumGetBoostsStatusRequest) TypeName() string {
	return "premium.getBoostsStatus"
}

// TypeInfo returns info about TL type.
func (g *PremiumGetBoostsStatusRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "premium.getBoostsStatus",
		ID:   PremiumGetBoostsStatusRequestTypeID,
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
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *PremiumGetBoostsStatusRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode premium.getBoostsStatus#42f1f61 as nil")
	}
	b.PutID(PremiumGetBoostsStatusRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PremiumGetBoostsStatusRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode premium.getBoostsStatus#42f1f61 as nil")
	}
	if g.Peer == nil {
		return fmt.Errorf("unable to encode premium.getBoostsStatus#42f1f61: field peer is nil")
	}
	if err := g.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode premium.getBoostsStatus#42f1f61: field peer: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *PremiumGetBoostsStatusRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode premium.getBoostsStatus#42f1f61 to nil")
	}
	if err := b.ConsumeID(PremiumGetBoostsStatusRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode premium.getBoostsStatus#42f1f61: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PremiumGetBoostsStatusRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode premium.getBoostsStatus#42f1f61 to nil")
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode premium.getBoostsStatus#42f1f61: field peer: %w", err)
		}
		g.Peer = value
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *PremiumGetBoostsStatusRequest) GetPeer() (value InputPeerClass) {
	if g == nil {
		return
	}
	return g.Peer
}

// PremiumGetBoostsStatus invokes method premium.getBoostsStatus#42f1f61 returning error if any.
//
// See https://core.telegram.org/method/premium.getBoostsStatus for reference.
func (c *Client) PremiumGetBoostsStatus(ctx context.Context, peer InputPeerClass) (*PremiumBoostsStatus, error) {
	var result PremiumBoostsStatus

	request := &PremiumGetBoostsStatusRequest{
		Peer: peer,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
