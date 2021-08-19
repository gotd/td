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

// ChannelsSetStickersRequest represents TL type `channels.setStickers#ea8ca4f9`.
// Associate a stickerset to the supergroup
//
// See https://core.telegram.org/method/channels.setStickers for reference.
type ChannelsSetStickersRequest struct {
	// Supergroup
	Channel InputChannelClass
	// The stickerset to associate
	Stickerset InputStickerSetClass
}

// ChannelsSetStickersRequestTypeID is TL type id of ChannelsSetStickersRequest.
const ChannelsSetStickersRequestTypeID = 0xea8ca4f9

func (s *ChannelsSetStickersRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Channel == nil) {
		return false
	}
	if !(s.Stickerset == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *ChannelsSetStickersRequest) String() string {
	if s == nil {
		return "ChannelsSetStickersRequest(nil)"
	}
	type Alias ChannelsSetStickersRequest
	return fmt.Sprintf("ChannelsSetStickersRequest%+v", Alias(*s))
}

// FillFrom fills ChannelsSetStickersRequest from given interface.
func (s *ChannelsSetStickersRequest) FillFrom(from interface {
	GetChannel() (value InputChannelClass)
	GetStickerset() (value InputStickerSetClass)
}) {
	s.Channel = from.GetChannel()
	s.Stickerset = from.GetStickerset()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsSetStickersRequest) TypeID() uint32 {
	return ChannelsSetStickersRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsSetStickersRequest) TypeName() string {
	return "channels.setStickers"
}

// TypeInfo returns info about TL type.
func (s *ChannelsSetStickersRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.setStickers",
		ID:   ChannelsSetStickersRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
		{
			Name:       "Stickerset",
			SchemaName: "stickerset",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *ChannelsSetStickersRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.setStickers#ea8ca4f9",
		}
	}
	b.PutID(ChannelsSetStickersRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *ChannelsSetStickersRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.setStickers#ea8ca4f9",
		}
	}
	if s.Channel == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.setStickers#ea8ca4f9",
			FieldName: "channel",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputChannel",
			},
		}
	}
	if err := s.Channel.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.setStickers#ea8ca4f9",
			FieldName:  "channel",
			Underlying: err,
		}
	}
	if s.Stickerset == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.setStickers#ea8ca4f9",
			FieldName: "stickerset",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputStickerSet",
			},
		}
	}
	if err := s.Stickerset.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.setStickers#ea8ca4f9",
			FieldName:  "stickerset",
			Underlying: err,
		}
	}
	return nil
}

// GetChannel returns value of Channel field.
func (s *ChannelsSetStickersRequest) GetChannel() (value InputChannelClass) {
	return s.Channel
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (s *ChannelsSetStickersRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return s.Channel.AsNotEmpty()
}

// GetStickerset returns value of Stickerset field.
func (s *ChannelsSetStickersRequest) GetStickerset() (value InputStickerSetClass) {
	return s.Stickerset
}

// Decode implements bin.Decoder.
func (s *ChannelsSetStickersRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.setStickers#ea8ca4f9",
		}
	}
	if err := b.ConsumeID(ChannelsSetStickersRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "channels.setStickers#ea8ca4f9",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *ChannelsSetStickersRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.setStickers#ea8ca4f9",
		}
	}
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.setStickers#ea8ca4f9",
				FieldName:  "channel",
				Underlying: err,
			}
		}
		s.Channel = value
	}
	{
		value, err := DecodeInputStickerSet(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.setStickers#ea8ca4f9",
				FieldName:  "stickerset",
				Underlying: err,
			}
		}
		s.Stickerset = value
	}
	return nil
}

// Ensuring interfaces in compile-time for ChannelsSetStickersRequest.
var (
	_ bin.Encoder     = &ChannelsSetStickersRequest{}
	_ bin.Decoder     = &ChannelsSetStickersRequest{}
	_ bin.BareEncoder = &ChannelsSetStickersRequest{}
	_ bin.BareDecoder = &ChannelsSetStickersRequest{}
)

// ChannelsSetStickers invokes method channels.setStickers#ea8ca4f9 returning error if any.
// Associate a stickerset to the supergroup
//
// Possible errors:
//  400 CHANNEL_INVALID: The provided channel is invalid
//  400 PARTICIPANTS_TOO_FEW: Not enough participants
//
// See https://core.telegram.org/method/channels.setStickers for reference.
// Can be used by bots.
func (c *Client) ChannelsSetStickers(ctx context.Context, request *ChannelsSetStickersRequest) (bool, error) {
	var result BoolBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
