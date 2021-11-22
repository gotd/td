// Code generated by gotdgen, DO NOT EDIT.

package tdapi

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

// GetInlineGameHighScoresRequest represents TL type `getInlineGameHighScores#92b7d658`.
type GetInlineGameHighScoresRequest struct {
	// Inline message identifier
	InlineMessageID string
	// User identifier
	UserID int32
}

// GetInlineGameHighScoresRequestTypeID is TL type id of GetInlineGameHighScoresRequest.
const GetInlineGameHighScoresRequestTypeID = 0x92b7d658

// Ensuring interfaces in compile-time for GetInlineGameHighScoresRequest.
var (
	_ bin.Encoder     = &GetInlineGameHighScoresRequest{}
	_ bin.Decoder     = &GetInlineGameHighScoresRequest{}
	_ bin.BareEncoder = &GetInlineGameHighScoresRequest{}
	_ bin.BareDecoder = &GetInlineGameHighScoresRequest{}
)

func (g *GetInlineGameHighScoresRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.InlineMessageID == "") {
		return false
	}
	if !(g.UserID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *GetInlineGameHighScoresRequest) String() string {
	if g == nil {
		return "GetInlineGameHighScoresRequest(nil)"
	}
	type Alias GetInlineGameHighScoresRequest
	return fmt.Sprintf("GetInlineGameHighScoresRequest%+v", Alias(*g))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*GetInlineGameHighScoresRequest) TypeID() uint32 {
	return GetInlineGameHighScoresRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*GetInlineGameHighScoresRequest) TypeName() string {
	return "getInlineGameHighScores"
}

// TypeInfo returns info about TL type.
func (g *GetInlineGameHighScoresRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "getInlineGameHighScores",
		ID:   GetInlineGameHighScoresRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "InlineMessageID",
			SchemaName: "inline_message_id",
		},
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *GetInlineGameHighScoresRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode getInlineGameHighScores#92b7d658 as nil")
	}
	b.PutID(GetInlineGameHighScoresRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *GetInlineGameHighScoresRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode getInlineGameHighScores#92b7d658 as nil")
	}
	b.PutString(g.InlineMessageID)
	b.PutInt32(g.UserID)
	return nil
}

// Decode implements bin.Decoder.
func (g *GetInlineGameHighScoresRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode getInlineGameHighScores#92b7d658 to nil")
	}
	if err := b.ConsumeID(GetInlineGameHighScoresRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode getInlineGameHighScores#92b7d658: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *GetInlineGameHighScoresRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode getInlineGameHighScores#92b7d658 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode getInlineGameHighScores#92b7d658: field inline_message_id: %w", err)
		}
		g.InlineMessageID = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode getInlineGameHighScores#92b7d658: field user_id: %w", err)
		}
		g.UserID = value
	}
	return nil
}

// GetInlineMessageID returns value of InlineMessageID field.
func (g *GetInlineGameHighScoresRequest) GetInlineMessageID() (value string) {
	return g.InlineMessageID
}

// GetUserID returns value of UserID field.
func (g *GetInlineGameHighScoresRequest) GetUserID() (value int32) {
	return g.UserID
}

// GetInlineGameHighScores invokes method getInlineGameHighScores#92b7d658 returning error if any.
func (c *Client) GetInlineGameHighScores(ctx context.Context, request *GetInlineGameHighScoresRequest) (*GameHighScores, error) {
	var result GameHighScores

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}