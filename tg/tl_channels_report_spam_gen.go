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

// ChannelsReportSpamRequest represents TL type `channels.reportSpam#fe087810`.
// Reports some messages from a user in a supergroup as spam; requires administrator
// rights in the supergroup
//
// See https://core.telegram.org/method/channels.reportSpam for reference.
type ChannelsReportSpamRequest struct {
	// Supergroup
	Channel InputChannelClass
	// ID of the user that sent the spam messages
	UserID InputUserClass
	// IDs of spam messages
	ID []int
}

// ChannelsReportSpamRequestTypeID is TL type id of ChannelsReportSpamRequest.
const ChannelsReportSpamRequestTypeID = 0xfe087810

func (r *ChannelsReportSpamRequest) Zero() bool {
	if r == nil {
		return true
	}
	if !(r.Channel == nil) {
		return false
	}
	if !(r.UserID == nil) {
		return false
	}
	if !(r.ID == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (r *ChannelsReportSpamRequest) String() string {
	if r == nil {
		return "ChannelsReportSpamRequest(nil)"
	}
	type Alias ChannelsReportSpamRequest
	return fmt.Sprintf("ChannelsReportSpamRequest%+v", Alias(*r))
}

// FillFrom fills ChannelsReportSpamRequest from given interface.
func (r *ChannelsReportSpamRequest) FillFrom(from interface {
	GetChannel() (value InputChannelClass)
	GetUserID() (value InputUserClass)
	GetID() (value []int)
}) {
	r.Channel = from.GetChannel()
	r.UserID = from.GetUserID()
	r.ID = from.GetID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsReportSpamRequest) TypeID() uint32 {
	return ChannelsReportSpamRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsReportSpamRequest) TypeName() string {
	return "channels.reportSpam"
}

// TypeInfo returns info about TL type.
func (r *ChannelsReportSpamRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.reportSpam",
		ID:   ChannelsReportSpamRequestTypeID,
	}
	if r == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
		{
			Name:       "ID",
			SchemaName: "id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (r *ChannelsReportSpamRequest) Encode(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.reportSpam#fe087810",
		}
	}
	b.PutID(ChannelsReportSpamRequestTypeID)
	return r.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (r *ChannelsReportSpamRequest) EncodeBare(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "channels.reportSpam#fe087810",
		}
	}
	if r.Channel == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.reportSpam#fe087810",
			FieldName: "channel",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputChannel",
			},
		}
	}
	if err := r.Channel.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.reportSpam#fe087810",
			FieldName:  "channel",
			Underlying: err,
		}
	}
	if r.UserID == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "channels.reportSpam#fe087810",
			FieldName: "user_id",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputUser",
			},
		}
	}
	if err := r.UserID.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "channels.reportSpam#fe087810",
			FieldName:  "user_id",
			Underlying: err,
		}
	}
	b.PutVectorHeader(len(r.ID))
	for _, v := range r.ID {
		b.PutInt(v)
	}
	return nil
}

// GetChannel returns value of Channel field.
func (r *ChannelsReportSpamRequest) GetChannel() (value InputChannelClass) {
	return r.Channel
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (r *ChannelsReportSpamRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return r.Channel.AsNotEmpty()
}

// GetUserID returns value of UserID field.
func (r *ChannelsReportSpamRequest) GetUserID() (value InputUserClass) {
	return r.UserID
}

// GetID returns value of ID field.
func (r *ChannelsReportSpamRequest) GetID() (value []int) {
	return r.ID
}

// Decode implements bin.Decoder.
func (r *ChannelsReportSpamRequest) Decode(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.reportSpam#fe087810",
		}
	}
	if err := b.ConsumeID(ChannelsReportSpamRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "channels.reportSpam#fe087810",
			Underlying: err,
		}
	}
	return r.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (r *ChannelsReportSpamRequest) DecodeBare(b *bin.Buffer) error {
	if r == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "channels.reportSpam#fe087810",
		}
	}
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.reportSpam#fe087810",
				FieldName:  "channel",
				Underlying: err,
			}
		}
		r.Channel = value
	}
	{
		value, err := DecodeInputUser(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.reportSpam#fe087810",
				FieldName:  "user_id",
				Underlying: err,
			}
		}
		r.UserID = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "channels.reportSpam#fe087810",
				FieldName:  "id",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			r.ID = make([]int, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.Int()
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "channels.reportSpam#fe087810",
					FieldName:  "id",
					Underlying: err,
				}
			}
			r.ID = append(r.ID, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for ChannelsReportSpamRequest.
var (
	_ bin.Encoder     = &ChannelsReportSpamRequest{}
	_ bin.Decoder     = &ChannelsReportSpamRequest{}
	_ bin.BareEncoder = &ChannelsReportSpamRequest{}
	_ bin.BareDecoder = &ChannelsReportSpamRequest{}
)

// ChannelsReportSpam invokes method channels.reportSpam#fe087810 returning error if any.
// Reports some messages from a user in a supergroup as spam; requires administrator
// rights in the supergroup
//
// Possible errors:
//  400 CHANNEL_INVALID: The provided channel is invalid
//  400 CHAT_ADMIN_REQUIRED: You must be an admin in this chat to do this
//  400 INPUT_USER_DEACTIVATED: The specified user was deleted
//  400 USER_ID_INVALID: The provided user ID is invalid
//
// See https://core.telegram.org/method/channels.reportSpam for reference.
func (c *Client) ChannelsReportSpam(ctx context.Context, request *ChannelsReportSpamRequest) (bool, error) {
	var result BoolBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
