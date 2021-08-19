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

// MessagesGetAdminsWithInvitesRequest represents TL type `messages.getAdminsWithInvites#3920e6ef`.
//
// See https://core.telegram.org/method/messages.getAdminsWithInvites for reference.
type MessagesGetAdminsWithInvitesRequest struct {
	// Peer field of MessagesGetAdminsWithInvitesRequest.
	Peer InputPeerClass
}

// MessagesGetAdminsWithInvitesRequestTypeID is TL type id of MessagesGetAdminsWithInvitesRequest.
const MessagesGetAdminsWithInvitesRequestTypeID = 0x3920e6ef

func (g *MessagesGetAdminsWithInvitesRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Peer == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *MessagesGetAdminsWithInvitesRequest) String() string {
	if g == nil {
		return "MessagesGetAdminsWithInvitesRequest(nil)"
	}
	type Alias MessagesGetAdminsWithInvitesRequest
	return fmt.Sprintf("MessagesGetAdminsWithInvitesRequest%+v", Alias(*g))
}

// FillFrom fills MessagesGetAdminsWithInvitesRequest from given interface.
func (g *MessagesGetAdminsWithInvitesRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
}) {
	g.Peer = from.GetPeer()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesGetAdminsWithInvitesRequest) TypeID() uint32 {
	return MessagesGetAdminsWithInvitesRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesGetAdminsWithInvitesRequest) TypeName() string {
	return "messages.getAdminsWithInvites"
}

// TypeInfo returns info about TL type.
func (g *MessagesGetAdminsWithInvitesRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.getAdminsWithInvites",
		ID:   MessagesGetAdminsWithInvitesRequestTypeID,
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
func (g *MessagesGetAdminsWithInvitesRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getAdminsWithInvites#3920e6ef",
		}
	}
	b.PutID(MessagesGetAdminsWithInvitesRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *MessagesGetAdminsWithInvitesRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getAdminsWithInvites#3920e6ef",
		}
	}
	if g.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "messages.getAdminsWithInvites#3920e6ef",
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
			TypeName:   "messages.getAdminsWithInvites#3920e6ef",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *MessagesGetAdminsWithInvitesRequest) GetPeer() (value InputPeerClass) {
	return g.Peer
}

// Decode implements bin.Decoder.
func (g *MessagesGetAdminsWithInvitesRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getAdminsWithInvites#3920e6ef",
		}
	}
	if err := b.ConsumeID(MessagesGetAdminsWithInvitesRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.getAdminsWithInvites#3920e6ef",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *MessagesGetAdminsWithInvitesRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getAdminsWithInvites#3920e6ef",
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.getAdminsWithInvites#3920e6ef",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		g.Peer = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesGetAdminsWithInvitesRequest.
var (
	_ bin.Encoder     = &MessagesGetAdminsWithInvitesRequest{}
	_ bin.Decoder     = &MessagesGetAdminsWithInvitesRequest{}
	_ bin.BareEncoder = &MessagesGetAdminsWithInvitesRequest{}
	_ bin.BareDecoder = &MessagesGetAdminsWithInvitesRequest{}
)

// MessagesGetAdminsWithInvites invokes method messages.getAdminsWithInvites#3920e6ef returning error if any.
//
// See https://core.telegram.org/method/messages.getAdminsWithInvites for reference.
func (c *Client) MessagesGetAdminsWithInvites(ctx context.Context, peer InputPeerClass) (*MessagesChatAdminsWithInvites, error) {
	var result MessagesChatAdminsWithInvites

	request := &MessagesGetAdminsWithInvitesRequest{
		Peer: peer,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
