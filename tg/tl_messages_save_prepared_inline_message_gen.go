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

// MessagesSavePreparedInlineMessageRequest represents TL type `messages.savePreparedInlineMessage#f21f7f2f`.
// Save a prepared inline message¹, to be shared by the user of the mini app using a
// web_app_send_prepared_message event²
//
// Links:
//  1. https://core.telegram.org/api/bots/inline#21-using-a-prepared-inline-message
//  2. https://core.telegram.org/api/web-events#web-app-send-prepared-message
//
// See https://core.telegram.org/method/messages.savePreparedInlineMessage for reference.
type MessagesSavePreparedInlineMessageRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// The message
	Result InputBotInlineResultClass
	// The user to whom the web_app_send_prepared_message event¹ event will be sent
	//
	// Links:
	//  1) https://core.telegram.org/api/web-events#web-app-send-prepared-message
	UserID InputUserClass
	// Types of chats where this message can be sent
	//
	// Use SetPeerTypes and GetPeerTypes helpers.
	PeerTypes []InlineQueryPeerTypeClass
}

// MessagesSavePreparedInlineMessageRequestTypeID is TL type id of MessagesSavePreparedInlineMessageRequest.
const MessagesSavePreparedInlineMessageRequestTypeID = 0xf21f7f2f

// Ensuring interfaces in compile-time for MessagesSavePreparedInlineMessageRequest.
var (
	_ bin.Encoder     = &MessagesSavePreparedInlineMessageRequest{}
	_ bin.Decoder     = &MessagesSavePreparedInlineMessageRequest{}
	_ bin.BareEncoder = &MessagesSavePreparedInlineMessageRequest{}
	_ bin.BareDecoder = &MessagesSavePreparedInlineMessageRequest{}
)

func (s *MessagesSavePreparedInlineMessageRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Flags.Zero()) {
		return false
	}
	if !(s.Result == nil) {
		return false
	}
	if !(s.UserID == nil) {
		return false
	}
	if !(s.PeerTypes == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *MessagesSavePreparedInlineMessageRequest) String() string {
	if s == nil {
		return "MessagesSavePreparedInlineMessageRequest(nil)"
	}
	type Alias MessagesSavePreparedInlineMessageRequest
	return fmt.Sprintf("MessagesSavePreparedInlineMessageRequest%+v", Alias(*s))
}

// FillFrom fills MessagesSavePreparedInlineMessageRequest from given interface.
func (s *MessagesSavePreparedInlineMessageRequest) FillFrom(from interface {
	GetResult() (value InputBotInlineResultClass)
	GetUserID() (value InputUserClass)
	GetPeerTypes() (value []InlineQueryPeerTypeClass, ok bool)
}) {
	s.Result = from.GetResult()
	s.UserID = from.GetUserID()
	if val, ok := from.GetPeerTypes(); ok {
		s.PeerTypes = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesSavePreparedInlineMessageRequest) TypeID() uint32 {
	return MessagesSavePreparedInlineMessageRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesSavePreparedInlineMessageRequest) TypeName() string {
	return "messages.savePreparedInlineMessage"
}

// TypeInfo returns info about TL type.
func (s *MessagesSavePreparedInlineMessageRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.savePreparedInlineMessage",
		ID:   MessagesSavePreparedInlineMessageRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Result",
			SchemaName: "result",
		},
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
		{
			Name:       "PeerTypes",
			SchemaName: "peer_types",
			Null:       !s.Flags.Has(0),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (s *MessagesSavePreparedInlineMessageRequest) SetFlags() {
	if !(s.PeerTypes == nil) {
		s.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (s *MessagesSavePreparedInlineMessageRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode messages.savePreparedInlineMessage#f21f7f2f as nil")
	}
	b.PutID(MessagesSavePreparedInlineMessageRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *MessagesSavePreparedInlineMessageRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode messages.savePreparedInlineMessage#f21f7f2f as nil")
	}
	s.SetFlags()
	if err := s.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field flags: %w", err)
	}
	if s.Result == nil {
		return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field result is nil")
	}
	if err := s.Result.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field result: %w", err)
	}
	if s.UserID == nil {
		return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field user_id is nil")
	}
	if err := s.UserID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field user_id: %w", err)
	}
	if s.Flags.Has(0) {
		b.PutVectorHeader(len(s.PeerTypes))
		for idx, v := range s.PeerTypes {
			if v == nil {
				return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field peer_types element with index %d is nil", idx)
			}
			if err := v.Encode(b); err != nil {
				return fmt.Errorf("unable to encode messages.savePreparedInlineMessage#f21f7f2f: field peer_types element with index %d: %w", idx, err)
			}
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *MessagesSavePreparedInlineMessageRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode messages.savePreparedInlineMessage#f21f7f2f to nil")
	}
	if err := b.ConsumeID(MessagesSavePreparedInlineMessageRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *MessagesSavePreparedInlineMessageRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode messages.savePreparedInlineMessage#f21f7f2f to nil")
	}
	{
		if err := s.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: field flags: %w", err)
		}
	}
	{
		value, err := DecodeInputBotInlineResult(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: field result: %w", err)
		}
		s.Result = value
	}
	{
		value, err := DecodeInputUser(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: field user_id: %w", err)
		}
		s.UserID = value
	}
	if s.Flags.Has(0) {
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: field peer_types: %w", err)
		}

		if headerLen > 0 {
			s.PeerTypes = make([]InlineQueryPeerTypeClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInlineQueryPeerType(b)
			if err != nil {
				return fmt.Errorf("unable to decode messages.savePreparedInlineMessage#f21f7f2f: field peer_types: %w", err)
			}
			s.PeerTypes = append(s.PeerTypes, value)
		}
	}
	return nil
}

// GetResult returns value of Result field.
func (s *MessagesSavePreparedInlineMessageRequest) GetResult() (value InputBotInlineResultClass) {
	if s == nil {
		return
	}
	return s.Result
}

// GetUserID returns value of UserID field.
func (s *MessagesSavePreparedInlineMessageRequest) GetUserID() (value InputUserClass) {
	if s == nil {
		return
	}
	return s.UserID
}

// SetPeerTypes sets value of PeerTypes conditional field.
func (s *MessagesSavePreparedInlineMessageRequest) SetPeerTypes(value []InlineQueryPeerTypeClass) {
	s.Flags.Set(0)
	s.PeerTypes = value
}

// GetPeerTypes returns value of PeerTypes conditional field and
// boolean which is true if field was set.
func (s *MessagesSavePreparedInlineMessageRequest) GetPeerTypes() (value []InlineQueryPeerTypeClass, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(0) {
		return value, false
	}
	return s.PeerTypes, true
}

// MapPeerTypes returns field PeerTypes wrapped in InlineQueryPeerTypeClassArray helper.
func (s *MessagesSavePreparedInlineMessageRequest) MapPeerTypes() (value InlineQueryPeerTypeClassArray, ok bool) {
	if !s.Flags.Has(0) {
		return value, false
	}
	return InlineQueryPeerTypeClassArray(s.PeerTypes), true
}

// MessagesSavePreparedInlineMessage invokes method messages.savePreparedInlineMessage#f21f7f2f returning error if any.
// Save a prepared inline message¹, to be shared by the user of the mini app using a
// web_app_send_prepared_message event²
//
// Links:
//  1. https://core.telegram.org/api/bots/inline#21-using-a-prepared-inline-message
//  2. https://core.telegram.org/api/web-events#web-app-send-prepared-message
//
// Possible errors:
//
//	400 RESULT_ID_INVALID: One of the specified result IDs is invalid.
//	400 USER_BOT_REQUIRED: This method can only be called by a bot.
//	400 USER_ID_INVALID: The provided user ID is invalid.
//
// See https://core.telegram.org/method/messages.savePreparedInlineMessage for reference.
// Can be used by bots.
func (c *Client) MessagesSavePreparedInlineMessage(ctx context.Context, request *MessagesSavePreparedInlineMessageRequest) (*MessagesBotPreparedInlineMessage, error) {
	var result MessagesBotPreparedInlineMessage

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
