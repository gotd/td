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

// SellGiftRequest represents TL type `sellGift#671b88b4`.
type SellGiftRequest struct {
	// Identifier of the user that sent the gift
	SenderUserID int64
	// Identifier of the message with the gift in the chat with the user
	MessageID int64
}

// SellGiftRequestTypeID is TL type id of SellGiftRequest.
const SellGiftRequestTypeID = 0x671b88b4

// Ensuring interfaces in compile-time for SellGiftRequest.
var (
	_ bin.Encoder     = &SellGiftRequest{}
	_ bin.Decoder     = &SellGiftRequest{}
	_ bin.BareEncoder = &SellGiftRequest{}
	_ bin.BareDecoder = &SellGiftRequest{}
)

func (s *SellGiftRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.SenderUserID == 0) {
		return false
	}
	if !(s.MessageID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SellGiftRequest) String() string {
	if s == nil {
		return "SellGiftRequest(nil)"
	}
	type Alias SellGiftRequest
	return fmt.Sprintf("SellGiftRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SellGiftRequest) TypeID() uint32 {
	return SellGiftRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SellGiftRequest) TypeName() string {
	return "sellGift"
}

// TypeInfo returns info about TL type.
func (s *SellGiftRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "sellGift",
		ID:   SellGiftRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "SenderUserID",
			SchemaName: "sender_user_id",
		},
		{
			Name:       "MessageID",
			SchemaName: "message_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SellGiftRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sellGift#671b88b4 as nil")
	}
	b.PutID(SellGiftRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SellGiftRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sellGift#671b88b4 as nil")
	}
	b.PutInt53(s.SenderUserID)
	b.PutInt53(s.MessageID)
	return nil
}

// Decode implements bin.Decoder.
func (s *SellGiftRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sellGift#671b88b4 to nil")
	}
	if err := b.ConsumeID(SellGiftRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode sellGift#671b88b4: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SellGiftRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sellGift#671b88b4 to nil")
	}
	{
		value, err := b.Int53()
		if err != nil {
			return fmt.Errorf("unable to decode sellGift#671b88b4: field sender_user_id: %w", err)
		}
		s.SenderUserID = value
	}
	{
		value, err := b.Int53()
		if err != nil {
			return fmt.Errorf("unable to decode sellGift#671b88b4: field message_id: %w", err)
		}
		s.MessageID = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (s *SellGiftRequest) EncodeTDLibJSON(b tdjson.Encoder) error {
	if s == nil {
		return fmt.Errorf("can't encode sellGift#671b88b4 as nil")
	}
	b.ObjStart()
	b.PutID("sellGift")
	b.Comma()
	b.FieldStart("sender_user_id")
	b.PutInt53(s.SenderUserID)
	b.Comma()
	b.FieldStart("message_id")
	b.PutInt53(s.MessageID)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (s *SellGiftRequest) DecodeTDLibJSON(b tdjson.Decoder) error {
	if s == nil {
		return fmt.Errorf("can't decode sellGift#671b88b4 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("sellGift"); err != nil {
				return fmt.Errorf("unable to decode sellGift#671b88b4: %w", err)
			}
		case "sender_user_id":
			value, err := b.Int53()
			if err != nil {
				return fmt.Errorf("unable to decode sellGift#671b88b4: field sender_user_id: %w", err)
			}
			s.SenderUserID = value
		case "message_id":
			value, err := b.Int53()
			if err != nil {
				return fmt.Errorf("unable to decode sellGift#671b88b4: field message_id: %w", err)
			}
			s.MessageID = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetSenderUserID returns value of SenderUserID field.
func (s *SellGiftRequest) GetSenderUserID() (value int64) {
	if s == nil {
		return
	}
	return s.SenderUserID
}

// GetMessageID returns value of MessageID field.
func (s *SellGiftRequest) GetMessageID() (value int64) {
	if s == nil {
		return
	}
	return s.MessageID
}

// SellGift invokes method sellGift#671b88b4 returning error if any.
func (c *Client) SellGift(ctx context.Context, request *SellGiftRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}
