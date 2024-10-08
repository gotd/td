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

// SetStoryReactionRequest represents TL type `setStoryReaction#ac8b4fa7`.
type SetStoryReactionRequest struct {
	// The identifier of the sender of the story
	StorySenderChatID int64
	// The identifier of the story
	StoryID int32
	// Type of the reaction to set; pass null to remove the reaction. Custom emoji reactions
	// can be used only by Telegram Premium users. Paid reactions can't be set
	ReactionType ReactionTypeClass
	// Pass true if the reaction needs to be added to recent reactions
	UpdateRecentReactions bool
}

// SetStoryReactionRequestTypeID is TL type id of SetStoryReactionRequest.
const SetStoryReactionRequestTypeID = 0xac8b4fa7

// Ensuring interfaces in compile-time for SetStoryReactionRequest.
var (
	_ bin.Encoder     = &SetStoryReactionRequest{}
	_ bin.Decoder     = &SetStoryReactionRequest{}
	_ bin.BareEncoder = &SetStoryReactionRequest{}
	_ bin.BareDecoder = &SetStoryReactionRequest{}
)

func (s *SetStoryReactionRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.StorySenderChatID == 0) {
		return false
	}
	if !(s.StoryID == 0) {
		return false
	}
	if !(s.ReactionType == nil) {
		return false
	}
	if !(s.UpdateRecentReactions == false) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SetStoryReactionRequest) String() string {
	if s == nil {
		return "SetStoryReactionRequest(nil)"
	}
	type Alias SetStoryReactionRequest
	return fmt.Sprintf("SetStoryReactionRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SetStoryReactionRequest) TypeID() uint32 {
	return SetStoryReactionRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SetStoryReactionRequest) TypeName() string {
	return "setStoryReaction"
}

// TypeInfo returns info about TL type.
func (s *SetStoryReactionRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "setStoryReaction",
		ID:   SetStoryReactionRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "StorySenderChatID",
			SchemaName: "story_sender_chat_id",
		},
		{
			Name:       "StoryID",
			SchemaName: "story_id",
		},
		{
			Name:       "ReactionType",
			SchemaName: "reaction_type",
		},
		{
			Name:       "UpdateRecentReactions",
			SchemaName: "update_recent_reactions",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SetStoryReactionRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setStoryReaction#ac8b4fa7 as nil")
	}
	b.PutID(SetStoryReactionRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SetStoryReactionRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setStoryReaction#ac8b4fa7 as nil")
	}
	b.PutInt53(s.StorySenderChatID)
	b.PutInt32(s.StoryID)
	if s.ReactionType == nil {
		return fmt.Errorf("unable to encode setStoryReaction#ac8b4fa7: field reaction_type is nil")
	}
	if err := s.ReactionType.Encode(b); err != nil {
		return fmt.Errorf("unable to encode setStoryReaction#ac8b4fa7: field reaction_type: %w", err)
	}
	b.PutBool(s.UpdateRecentReactions)
	return nil
}

// Decode implements bin.Decoder.
func (s *SetStoryReactionRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setStoryReaction#ac8b4fa7 to nil")
	}
	if err := b.ConsumeID(SetStoryReactionRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SetStoryReactionRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setStoryReaction#ac8b4fa7 to nil")
	}
	{
		value, err := b.Int53()
		if err != nil {
			return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field story_sender_chat_id: %w", err)
		}
		s.StorySenderChatID = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field story_id: %w", err)
		}
		s.StoryID = value
	}
	{
		value, err := DecodeReactionType(b)
		if err != nil {
			return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field reaction_type: %w", err)
		}
		s.ReactionType = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field update_recent_reactions: %w", err)
		}
		s.UpdateRecentReactions = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (s *SetStoryReactionRequest) EncodeTDLibJSON(b tdjson.Encoder) error {
	if s == nil {
		return fmt.Errorf("can't encode setStoryReaction#ac8b4fa7 as nil")
	}
	b.ObjStart()
	b.PutID("setStoryReaction")
	b.Comma()
	b.FieldStart("story_sender_chat_id")
	b.PutInt53(s.StorySenderChatID)
	b.Comma()
	b.FieldStart("story_id")
	b.PutInt32(s.StoryID)
	b.Comma()
	b.FieldStart("reaction_type")
	if s.ReactionType == nil {
		return fmt.Errorf("unable to encode setStoryReaction#ac8b4fa7: field reaction_type is nil")
	}
	if err := s.ReactionType.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode setStoryReaction#ac8b4fa7: field reaction_type: %w", err)
	}
	b.Comma()
	b.FieldStart("update_recent_reactions")
	b.PutBool(s.UpdateRecentReactions)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (s *SetStoryReactionRequest) DecodeTDLibJSON(b tdjson.Decoder) error {
	if s == nil {
		return fmt.Errorf("can't decode setStoryReaction#ac8b4fa7 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("setStoryReaction"); err != nil {
				return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: %w", err)
			}
		case "story_sender_chat_id":
			value, err := b.Int53()
			if err != nil {
				return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field story_sender_chat_id: %w", err)
			}
			s.StorySenderChatID = value
		case "story_id":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field story_id: %w", err)
			}
			s.StoryID = value
		case "reaction_type":
			value, err := DecodeTDLibJSONReactionType(b)
			if err != nil {
				return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field reaction_type: %w", err)
			}
			s.ReactionType = value
		case "update_recent_reactions":
			value, err := b.Bool()
			if err != nil {
				return fmt.Errorf("unable to decode setStoryReaction#ac8b4fa7: field update_recent_reactions: %w", err)
			}
			s.UpdateRecentReactions = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetStorySenderChatID returns value of StorySenderChatID field.
func (s *SetStoryReactionRequest) GetStorySenderChatID() (value int64) {
	if s == nil {
		return
	}
	return s.StorySenderChatID
}

// GetStoryID returns value of StoryID field.
func (s *SetStoryReactionRequest) GetStoryID() (value int32) {
	if s == nil {
		return
	}
	return s.StoryID
}

// GetReactionType returns value of ReactionType field.
func (s *SetStoryReactionRequest) GetReactionType() (value ReactionTypeClass) {
	if s == nil {
		return
	}
	return s.ReactionType
}

// GetUpdateRecentReactions returns value of UpdateRecentReactions field.
func (s *SetStoryReactionRequest) GetUpdateRecentReactions() (value bool) {
	if s == nil {
		return
	}
	return s.UpdateRecentReactions
}

// SetStoryReaction invokes method setStoryReaction#ac8b4fa7 returning error if any.
func (c *Client) SetStoryReaction(ctx context.Context, request *SetStoryReactionRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}
