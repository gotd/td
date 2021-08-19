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

// ChannelsGetParticipantRequest represents TL type `channels.getParticipant#a0ab6cc6`.
// Get info about a channel/supergroup¹ participant
//
// Links:
//  1) https://core.telegram.org/api/channel
//
// See https://core.telegram.org/method/channels.getParticipant for reference.
type ChannelsGetParticipantRequest struct {
	// Channel/supergroup
	Channel InputChannelClass
	// Participant field of ChannelsGetParticipantRequest.
	Participant InputPeerClass
}

// ChannelsGetParticipantRequestTypeID is TL type id of ChannelsGetParticipantRequest.
const ChannelsGetParticipantRequestTypeID = 0xa0ab6cc6

func (g *ChannelsGetParticipantRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Channel == nil) {
		return false
	}
	if !(g.Participant == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *ChannelsGetParticipantRequest) String() string {
	if g == nil {
		return "ChannelsGetParticipantRequest(nil)"
	}
	type Alias ChannelsGetParticipantRequest
	return fmt.Sprintf("ChannelsGetParticipantRequest%+v", Alias(*g))
}

// FillFrom fills ChannelsGetParticipantRequest from given interface.
func (g *ChannelsGetParticipantRequest) FillFrom(from interface {
	GetChannel() (value InputChannelClass)
	GetParticipant() (value InputPeerClass)
}) {
	g.Channel = from.GetChannel()
	g.Participant = from.GetParticipant()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsGetParticipantRequest) TypeID() uint32 {
	return ChannelsGetParticipantRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsGetParticipantRequest) TypeName() string {
	return "channels.getParticipant"
}

// TypeInfo returns info about TL type.
func (g *ChannelsGetParticipantRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.getParticipant",
		ID:   ChannelsGetParticipantRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
		{
			Name:       "Participant",
			SchemaName: "participant",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *ChannelsGetParticipantRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.getParticipant#a0ab6cc6",
		}
	}
	b.PutID(ChannelsGetParticipantRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *ChannelsGetParticipantRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.getParticipant#a0ab6cc6",
		}
	}
	if g.Channel == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.getParticipant#a0ab6cc6",
			FieldName: "channel",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputChannel",
			},
		}
	}
	if err := g.Channel.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.getParticipant#a0ab6cc6",
			FieldName:  "channel",
			Underlying: err,
		}
	}
	if g.Participant == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.getParticipant#a0ab6cc6",
			FieldName: "participant",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputPeer",
			},
		}
	}
	if err := g.Participant.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.getParticipant#a0ab6cc6",
			FieldName:  "participant",
			Underlying: err,
		}
	}
	return nil
}

// GetChannel returns value of Channel field.
func (g *ChannelsGetParticipantRequest) GetChannel() (value InputChannelClass) {
	return g.Channel
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (g *ChannelsGetParticipantRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return g.Channel.AsNotEmpty()
}

// GetParticipant returns value of Participant field.
func (g *ChannelsGetParticipantRequest) GetParticipant() (value InputPeerClass) {
	return g.Participant
}

// Decode implements bin.Decoder.
func (g *ChannelsGetParticipantRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.getParticipant#a0ab6cc6",
		}
	}
	if err := b.ConsumeID(ChannelsGetParticipantRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "channels.getParticipant#a0ab6cc6",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *ChannelsGetParticipantRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.getParticipant#a0ab6cc6",
		}
	}
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.getParticipant#a0ab6cc6",
				FieldName:  "channel",
				Underlying: err,
			}
		}
		g.Channel = value
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.getParticipant#a0ab6cc6",
				FieldName:  "participant",
				Underlying: err,
			}
		}
		g.Participant = value
	}
	return nil
}

// Ensuring interfaces in compile-time for ChannelsGetParticipantRequest.
var (
	_ bin.Encoder     = &ChannelsGetParticipantRequest{}
	_ bin.Decoder     = &ChannelsGetParticipantRequest{}
	_ bin.BareEncoder = &ChannelsGetParticipantRequest{}
	_ bin.BareDecoder = &ChannelsGetParticipantRequest{}
)

// ChannelsGetParticipant invokes method channels.getParticipant#a0ab6cc6 returning error if any.
// Get info about a channel/supergroup¹ participant
//
// Links:
//  1) https://core.telegram.org/api/channel
//
// Possible errors:
//  400 CHANNEL_INVALID: The provided channel is invalid
//  400 CHANNEL_PRIVATE: You haven't joined this channel/supergroup
//  400 CHAT_ADMIN_REQUIRED: You must be an admin in this chat to do this
//  400 MSG_ID_INVALID: Invalid message ID provided
//  400 USER_ID_INVALID: The provided user ID is invalid
//  400 USER_NOT_PARTICIPANT: You're not a member of this supergroup/channel
//
// See https://core.telegram.org/method/channels.getParticipant for reference.
// Can be used by bots.
func (c *Client) ChannelsGetParticipant(ctx context.Context, request *ChannelsGetParticipantRequest) (*ChannelsChannelParticipant, error) {
	var result ChannelsChannelParticipant

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
