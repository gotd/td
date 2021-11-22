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

// SendMessageAlbumRequest represents TL type `sendMessageAlbum#ae6f51e6`.
type SendMessageAlbumRequest struct {
	// Target chat
	ChatID int64
	// If not 0, a message thread identifier in which the messages will be sent
	MessageThreadID int64
	// Identifier of a message to reply to or 0
	ReplyToMessageID int64
	// Options to be used to send the messages
	Options MessageSendOptions
	// Contents of messages to be sent. At most 10 messages can be added to an album
	InputMessageContents []InputMessageContentClass
}

// SendMessageAlbumRequestTypeID is TL type id of SendMessageAlbumRequest.
const SendMessageAlbumRequestTypeID = 0xae6f51e6

// Ensuring interfaces in compile-time for SendMessageAlbumRequest.
var (
	_ bin.Encoder     = &SendMessageAlbumRequest{}
	_ bin.Decoder     = &SendMessageAlbumRequest{}
	_ bin.BareEncoder = &SendMessageAlbumRequest{}
	_ bin.BareDecoder = &SendMessageAlbumRequest{}
)

func (s *SendMessageAlbumRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.ChatID == 0) {
		return false
	}
	if !(s.MessageThreadID == 0) {
		return false
	}
	if !(s.ReplyToMessageID == 0) {
		return false
	}
	if !(s.Options.Zero()) {
		return false
	}
	if !(s.InputMessageContents == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SendMessageAlbumRequest) String() string {
	if s == nil {
		return "SendMessageAlbumRequest(nil)"
	}
	type Alias SendMessageAlbumRequest
	return fmt.Sprintf("SendMessageAlbumRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SendMessageAlbumRequest) TypeID() uint32 {
	return SendMessageAlbumRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SendMessageAlbumRequest) TypeName() string {
	return "sendMessageAlbum"
}

// TypeInfo returns info about TL type.
func (s *SendMessageAlbumRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "sendMessageAlbum",
		ID:   SendMessageAlbumRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ChatID",
			SchemaName: "chat_id",
		},
		{
			Name:       "MessageThreadID",
			SchemaName: "message_thread_id",
		},
		{
			Name:       "ReplyToMessageID",
			SchemaName: "reply_to_message_id",
		},
		{
			Name:       "Options",
			SchemaName: "options",
		},
		{
			Name:       "InputMessageContents",
			SchemaName: "input_message_contents",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SendMessageAlbumRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sendMessageAlbum#ae6f51e6 as nil")
	}
	b.PutID(SendMessageAlbumRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SendMessageAlbumRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sendMessageAlbum#ae6f51e6 as nil")
	}
	b.PutLong(s.ChatID)
	b.PutLong(s.MessageThreadID)
	b.PutLong(s.ReplyToMessageID)
	if err := s.Options.Encode(b); err != nil {
		return fmt.Errorf("unable to encode sendMessageAlbum#ae6f51e6: field options: %w", err)
	}
	b.PutInt(len(s.InputMessageContents))
	for idx, v := range s.InputMessageContents {
		if v == nil {
			return fmt.Errorf("unable to encode sendMessageAlbum#ae6f51e6: field input_message_contents element with index %d is nil", idx)
		}
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare sendMessageAlbum#ae6f51e6: field input_message_contents element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SendMessageAlbumRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sendMessageAlbum#ae6f51e6 to nil")
	}
	if err := b.ConsumeID(SendMessageAlbumRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SendMessageAlbumRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sendMessageAlbum#ae6f51e6 to nil")
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field chat_id: %w", err)
		}
		s.ChatID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field message_thread_id: %w", err)
		}
		s.MessageThreadID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field reply_to_message_id: %w", err)
		}
		s.ReplyToMessageID = value
	}
	{
		if err := s.Options.Decode(b); err != nil {
			return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field options: %w", err)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field input_message_contents: %w", err)
		}

		if headerLen > 0 {
			s.InputMessageContents = make([]InputMessageContentClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputMessageContent(b)
			if err != nil {
				return fmt.Errorf("unable to decode sendMessageAlbum#ae6f51e6: field input_message_contents: %w", err)
			}
			s.InputMessageContents = append(s.InputMessageContents, value)
		}
	}
	return nil
}

// GetChatID returns value of ChatID field.
func (s *SendMessageAlbumRequest) GetChatID() (value int64) {
	return s.ChatID
}

// GetMessageThreadID returns value of MessageThreadID field.
func (s *SendMessageAlbumRequest) GetMessageThreadID() (value int64) {
	return s.MessageThreadID
}

// GetReplyToMessageID returns value of ReplyToMessageID field.
func (s *SendMessageAlbumRequest) GetReplyToMessageID() (value int64) {
	return s.ReplyToMessageID
}

// GetOptions returns value of Options field.
func (s *SendMessageAlbumRequest) GetOptions() (value MessageSendOptions) {
	return s.Options
}

// GetInputMessageContents returns value of InputMessageContents field.
func (s *SendMessageAlbumRequest) GetInputMessageContents() (value []InputMessageContentClass) {
	return s.InputMessageContents
}

// SendMessageAlbum invokes method sendMessageAlbum#ae6f51e6 returning error if any.
func (c *Client) SendMessageAlbum(ctx context.Context, request *SendMessageAlbumRequest) (*Messages, error) {
	var result Messages

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}