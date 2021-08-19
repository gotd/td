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

// PhoneGetGroupCallJoinAsRequest represents TL type `phone.getGroupCallJoinAs#ef7c213a`.
//
// See https://core.telegram.org/method/phone.getGroupCallJoinAs for reference.
type PhoneGetGroupCallJoinAsRequest struct {
	// Peer field of PhoneGetGroupCallJoinAsRequest.
	Peer InputPeerClass
}

// PhoneGetGroupCallJoinAsRequestTypeID is TL type id of PhoneGetGroupCallJoinAsRequest.
const PhoneGetGroupCallJoinAsRequestTypeID = 0xef7c213a

func (g *PhoneGetGroupCallJoinAsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Peer == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *PhoneGetGroupCallJoinAsRequest) String() string {
	if g == nil {
		return "PhoneGetGroupCallJoinAsRequest(nil)"
	}
	type Alias PhoneGetGroupCallJoinAsRequest
	return fmt.Sprintf("PhoneGetGroupCallJoinAsRequest%+v", Alias(*g))
}

// FillFrom fills PhoneGetGroupCallJoinAsRequest from given interface.
func (g *PhoneGetGroupCallJoinAsRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
}) {
	g.Peer = from.GetPeer()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PhoneGetGroupCallJoinAsRequest) TypeID() uint32 {
	return PhoneGetGroupCallJoinAsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PhoneGetGroupCallJoinAsRequest) TypeName() string {
	return "phone.getGroupCallJoinAs"
}

// TypeInfo returns info about TL type.
func (g *PhoneGetGroupCallJoinAsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "phone.getGroupCallJoinAs",
		ID:   PhoneGetGroupCallJoinAsRequestTypeID,
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
func (g *PhoneGetGroupCallJoinAsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.getGroupCallJoinAs#ef7c213a",
		}
	}
	b.PutID(PhoneGetGroupCallJoinAsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PhoneGetGroupCallJoinAsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "phone.getGroupCallJoinAs#ef7c213a",
		}
	}
	if g.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "phone.getGroupCallJoinAs#ef7c213a",
			FieldName: "peer",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputPeer",
			},
		}
	}
	if err := g.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "phone.getGroupCallJoinAs#ef7c213a",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *PhoneGetGroupCallJoinAsRequest) GetPeer() (value InputPeerClass) {
	return g.Peer
}

// Decode implements bin.Decoder.
func (g *PhoneGetGroupCallJoinAsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.getGroupCallJoinAs#ef7c213a",
		}
	}
	if err := b.ConsumeID(PhoneGetGroupCallJoinAsRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "phone.getGroupCallJoinAs#ef7c213a",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PhoneGetGroupCallJoinAsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "phone.getGroupCallJoinAs#ef7c213a",
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "phone.getGroupCallJoinAs#ef7c213a",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		g.Peer = value
	}
	return nil
}

// Ensuring interfaces in compile-time for PhoneGetGroupCallJoinAsRequest.
var (
	_ bin.Encoder     = &PhoneGetGroupCallJoinAsRequest{}
	_ bin.Decoder     = &PhoneGetGroupCallJoinAsRequest{}
	_ bin.BareEncoder = &PhoneGetGroupCallJoinAsRequest{}
	_ bin.BareDecoder = &PhoneGetGroupCallJoinAsRequest{}
)

// PhoneGetGroupCallJoinAs invokes method phone.getGroupCallJoinAs#ef7c213a returning error if any.
//
// See https://core.telegram.org/method/phone.getGroupCallJoinAs for reference.
func (c *Client) PhoneGetGroupCallJoinAs(ctx context.Context, peer InputPeerClass) (*PhoneJoinAsPeers, error) {
	var result PhoneJoinAsPeers

	request := &PhoneGetGroupCallJoinAsRequest{
		Peer: peer,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
