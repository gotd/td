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

// MessagesClearAllDraftsRequest represents TL type `messages.clearAllDrafts#7e58ee9c`.
// Clear all drafts¹.
//
// Links:
//  1) https://core.telegram.org/api/drafts
//
// See https://core.telegram.org/method/messages.clearAllDrafts for reference.
type MessagesClearAllDraftsRequest struct {
}

// MessagesClearAllDraftsRequestTypeID is TL type id of MessagesClearAllDraftsRequest.
const MessagesClearAllDraftsRequestTypeID = 0x7e58ee9c

func (c *MessagesClearAllDraftsRequest) Zero() bool {
	if c == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (c *MessagesClearAllDraftsRequest) String() string {
	if c == nil {
		return "MessagesClearAllDraftsRequest(nil)"
	}
	type Alias MessagesClearAllDraftsRequest
	return fmt.Sprintf("MessagesClearAllDraftsRequest%+v", Alias(*c))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesClearAllDraftsRequest) TypeID() uint32 {
	return MessagesClearAllDraftsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesClearAllDraftsRequest) TypeName() string {
	return "messages.clearAllDrafts"
}

// TypeInfo returns info about TL type.
func (c *MessagesClearAllDraftsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.clearAllDrafts",
		ID:   MessagesClearAllDraftsRequestTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (c *MessagesClearAllDraftsRequest) Encode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.clearAllDrafts#7e58ee9c",
		}
	}
	b.PutID(MessagesClearAllDraftsRequestTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *MessagesClearAllDraftsRequest) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.clearAllDrafts#7e58ee9c",
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (c *MessagesClearAllDraftsRequest) Decode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.clearAllDrafts#7e58ee9c",
		}
	}
	if err := b.ConsumeID(MessagesClearAllDraftsRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.clearAllDrafts#7e58ee9c",
			Underlying: err,
		}
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *MessagesClearAllDraftsRequest) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.clearAllDrafts#7e58ee9c",
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesClearAllDraftsRequest.
var (
	_ bin.Encoder     = &MessagesClearAllDraftsRequest{}
	_ bin.Decoder     = &MessagesClearAllDraftsRequest{}
	_ bin.BareEncoder = &MessagesClearAllDraftsRequest{}
	_ bin.BareDecoder = &MessagesClearAllDraftsRequest{}
)

// MessagesClearAllDrafts invokes method messages.clearAllDrafts#7e58ee9c returning error if any.
// Clear all drafts¹.
//
// Links:
//  1) https://core.telegram.org/api/drafts
//
// See https://core.telegram.org/method/messages.clearAllDrafts for reference.
func (c *Client) MessagesClearAllDrafts(ctx context.Context) (bool, error) {
	var result BoolBox

	request := &MessagesClearAllDraftsRequest{}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
