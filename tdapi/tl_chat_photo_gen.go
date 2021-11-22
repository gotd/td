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

// ChatPhoto represents TL type `chatPhoto#77176e42`.
type ChatPhoto struct {
	// Unique photo identifier
	ID Int64
	// Point in time (Unix timestamp) when the photo has been added
	AddedDate int32
	// Photo minithumbnail; may be null
	Minithumbnail Minithumbnail
	// Available variants of the photo in JPEG format, in different size
	Sizes []PhotoSize
	// Animated variant of the photo in MPEG4 format; may be null
	Animation AnimatedChatPhoto
}

// ChatPhotoTypeID is TL type id of ChatPhoto.
const ChatPhotoTypeID = 0x77176e42

// Ensuring interfaces in compile-time for ChatPhoto.
var (
	_ bin.Encoder     = &ChatPhoto{}
	_ bin.Decoder     = &ChatPhoto{}
	_ bin.BareEncoder = &ChatPhoto{}
	_ bin.BareDecoder = &ChatPhoto{}
)

func (c *ChatPhoto) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.ID.Zero()) {
		return false
	}
	if !(c.AddedDate == 0) {
		return false
	}
	if !(c.Minithumbnail.Zero()) {
		return false
	}
	if !(c.Sizes == nil) {
		return false
	}
	if !(c.Animation.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *ChatPhoto) String() string {
	if c == nil {
		return "ChatPhoto(nil)"
	}
	type Alias ChatPhoto
	return fmt.Sprintf("ChatPhoto%+v", Alias(*c))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChatPhoto) TypeID() uint32 {
	return ChatPhotoTypeID
}

// TypeName returns name of type in TL schema.
func (*ChatPhoto) TypeName() string {
	return "chatPhoto"
}

// TypeInfo returns info about TL type.
func (c *ChatPhoto) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "chatPhoto",
		ID:   ChatPhotoTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ID",
			SchemaName: "id",
		},
		{
			Name:       "AddedDate",
			SchemaName: "added_date",
		},
		{
			Name:       "Minithumbnail",
			SchemaName: "minithumbnail",
		},
		{
			Name:       "Sizes",
			SchemaName: "sizes",
		},
		{
			Name:       "Animation",
			SchemaName: "animation",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *ChatPhoto) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode chatPhoto#77176e42 as nil")
	}
	b.PutID(ChatPhotoTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *ChatPhoto) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode chatPhoto#77176e42 as nil")
	}
	if err := c.ID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode chatPhoto#77176e42: field id: %w", err)
	}
	b.PutInt32(c.AddedDate)
	if err := c.Minithumbnail.Encode(b); err != nil {
		return fmt.Errorf("unable to encode chatPhoto#77176e42: field minithumbnail: %w", err)
	}
	b.PutInt(len(c.Sizes))
	for idx, v := range c.Sizes {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare chatPhoto#77176e42: field sizes element with index %d: %w", idx, err)
		}
	}
	if err := c.Animation.Encode(b); err != nil {
		return fmt.Errorf("unable to encode chatPhoto#77176e42: field animation: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (c *ChatPhoto) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode chatPhoto#77176e42 to nil")
	}
	if err := b.ConsumeID(ChatPhotoTypeID); err != nil {
		return fmt.Errorf("unable to decode chatPhoto#77176e42: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *ChatPhoto) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode chatPhoto#77176e42 to nil")
	}
	{
		if err := c.ID.Decode(b); err != nil {
			return fmt.Errorf("unable to decode chatPhoto#77176e42: field id: %w", err)
		}
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode chatPhoto#77176e42: field added_date: %w", err)
		}
		c.AddedDate = value
	}
	{
		if err := c.Minithumbnail.Decode(b); err != nil {
			return fmt.Errorf("unable to decode chatPhoto#77176e42: field minithumbnail: %w", err)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode chatPhoto#77176e42: field sizes: %w", err)
		}

		if headerLen > 0 {
			c.Sizes = make([]PhotoSize, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value PhotoSize
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare chatPhoto#77176e42: field sizes: %w", err)
			}
			c.Sizes = append(c.Sizes, value)
		}
	}
	{
		if err := c.Animation.Decode(b); err != nil {
			return fmt.Errorf("unable to decode chatPhoto#77176e42: field animation: %w", err)
		}
	}
	return nil
}

// GetID returns value of ID field.
func (c *ChatPhoto) GetID() (value Int64) {
	return c.ID
}

// GetAddedDate returns value of AddedDate field.
func (c *ChatPhoto) GetAddedDate() (value int32) {
	return c.AddedDate
}

// GetMinithumbnail returns value of Minithumbnail field.
func (c *ChatPhoto) GetMinithumbnail() (value Minithumbnail) {
	return c.Minithumbnail
}

// GetSizes returns value of Sizes field.
func (c *ChatPhoto) GetSizes() (value []PhotoSize) {
	return c.Sizes
}

// GetAnimation returns value of Animation field.
func (c *ChatPhoto) GetAnimation() (value AnimatedChatPhoto) {
	return c.Animation
}