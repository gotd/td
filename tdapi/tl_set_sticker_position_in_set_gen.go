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

// SetStickerPositionInSetRequest represents TL type `setStickerPositionInSet#7bb24721`.
type SetStickerPositionInSetRequest struct {
	// Sticker
	Sticker InputFileClass
	// New position of the sticker in the set, zero-based
	Position int32
}

// SetStickerPositionInSetRequestTypeID is TL type id of SetStickerPositionInSetRequest.
const SetStickerPositionInSetRequestTypeID = 0x7bb24721

// Ensuring interfaces in compile-time for SetStickerPositionInSetRequest.
var (
	_ bin.Encoder     = &SetStickerPositionInSetRequest{}
	_ bin.Decoder     = &SetStickerPositionInSetRequest{}
	_ bin.BareEncoder = &SetStickerPositionInSetRequest{}
	_ bin.BareDecoder = &SetStickerPositionInSetRequest{}
)

func (s *SetStickerPositionInSetRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Sticker == nil) {
		return false
	}
	if !(s.Position == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SetStickerPositionInSetRequest) String() string {
	if s == nil {
		return "SetStickerPositionInSetRequest(nil)"
	}
	type Alias SetStickerPositionInSetRequest
	return fmt.Sprintf("SetStickerPositionInSetRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SetStickerPositionInSetRequest) TypeID() uint32 {
	return SetStickerPositionInSetRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SetStickerPositionInSetRequest) TypeName() string {
	return "setStickerPositionInSet"
}

// TypeInfo returns info about TL type.
func (s *SetStickerPositionInSetRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "setStickerPositionInSet",
		ID:   SetStickerPositionInSetRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Sticker",
			SchemaName: "sticker",
		},
		{
			Name:       "Position",
			SchemaName: "position",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SetStickerPositionInSetRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setStickerPositionInSet#7bb24721 as nil")
	}
	b.PutID(SetStickerPositionInSetRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SetStickerPositionInSetRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setStickerPositionInSet#7bb24721 as nil")
	}
	if s.Sticker == nil {
		return fmt.Errorf("unable to encode setStickerPositionInSet#7bb24721: field sticker is nil")
	}
	if err := s.Sticker.Encode(b); err != nil {
		return fmt.Errorf("unable to encode setStickerPositionInSet#7bb24721: field sticker: %w", err)
	}
	b.PutInt32(s.Position)
	return nil
}

// Decode implements bin.Decoder.
func (s *SetStickerPositionInSetRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setStickerPositionInSet#7bb24721 to nil")
	}
	if err := b.ConsumeID(SetStickerPositionInSetRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode setStickerPositionInSet#7bb24721: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SetStickerPositionInSetRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setStickerPositionInSet#7bb24721 to nil")
	}
	{
		value, err := DecodeInputFile(b)
		if err != nil {
			return fmt.Errorf("unable to decode setStickerPositionInSet#7bb24721: field sticker: %w", err)
		}
		s.Sticker = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode setStickerPositionInSet#7bb24721: field position: %w", err)
		}
		s.Position = value
	}
	return nil
}

// GetSticker returns value of Sticker field.
func (s *SetStickerPositionInSetRequest) GetSticker() (value InputFileClass) {
	return s.Sticker
}

// GetPosition returns value of Position field.
func (s *SetStickerPositionInSetRequest) GetPosition() (value int32) {
	return s.Position
}

// SetStickerPositionInSet invokes method setStickerPositionInSet#7bb24721 returning error if any.
func (c *Client) SetStickerPositionInSet(ctx context.Context, request *SetStickerPositionInSetRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}