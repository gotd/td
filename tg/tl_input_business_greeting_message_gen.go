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

// InputBusinessGreetingMessage represents TL type `inputBusinessGreetingMessage#194cb3b`.
//
// See https://core.telegram.org/constructor/inputBusinessGreetingMessage for reference.
type InputBusinessGreetingMessage struct {
	// ShortcutID field of InputBusinessGreetingMessage.
	ShortcutID int
	// Recipients field of InputBusinessGreetingMessage.
	Recipients InputBusinessRecipients
	// NoActivityDays field of InputBusinessGreetingMessage.
	NoActivityDays int
}

// InputBusinessGreetingMessageTypeID is TL type id of InputBusinessGreetingMessage.
const InputBusinessGreetingMessageTypeID = 0x194cb3b

// Ensuring interfaces in compile-time for InputBusinessGreetingMessage.
var (
	_ bin.Encoder     = &InputBusinessGreetingMessage{}
	_ bin.Decoder     = &InputBusinessGreetingMessage{}
	_ bin.BareEncoder = &InputBusinessGreetingMessage{}
	_ bin.BareDecoder = &InputBusinessGreetingMessage{}
)

func (i *InputBusinessGreetingMessage) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.ShortcutID == 0) {
		return false
	}
	if !(i.Recipients.Zero()) {
		return false
	}
	if !(i.NoActivityDays == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputBusinessGreetingMessage) String() string {
	if i == nil {
		return "InputBusinessGreetingMessage(nil)"
	}
	type Alias InputBusinessGreetingMessage
	return fmt.Sprintf("InputBusinessGreetingMessage%+v", Alias(*i))
}

// FillFrom fills InputBusinessGreetingMessage from given interface.
func (i *InputBusinessGreetingMessage) FillFrom(from interface {
	GetShortcutID() (value int)
	GetRecipients() (value InputBusinessRecipients)
	GetNoActivityDays() (value int)
}) {
	i.ShortcutID = from.GetShortcutID()
	i.Recipients = from.GetRecipients()
	i.NoActivityDays = from.GetNoActivityDays()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputBusinessGreetingMessage) TypeID() uint32 {
	return InputBusinessGreetingMessageTypeID
}

// TypeName returns name of type in TL schema.
func (*InputBusinessGreetingMessage) TypeName() string {
	return "inputBusinessGreetingMessage"
}

// TypeInfo returns info about TL type.
func (i *InputBusinessGreetingMessage) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputBusinessGreetingMessage",
		ID:   InputBusinessGreetingMessageTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ShortcutID",
			SchemaName: "shortcut_id",
		},
		{
			Name:       "Recipients",
			SchemaName: "recipients",
		},
		{
			Name:       "NoActivityDays",
			SchemaName: "no_activity_days",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputBusinessGreetingMessage) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputBusinessGreetingMessage#194cb3b as nil")
	}
	b.PutID(InputBusinessGreetingMessageTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputBusinessGreetingMessage) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputBusinessGreetingMessage#194cb3b as nil")
	}
	b.PutInt(i.ShortcutID)
	if err := i.Recipients.Encode(b); err != nil {
		return fmt.Errorf("unable to encode inputBusinessGreetingMessage#194cb3b: field recipients: %w", err)
	}
	b.PutInt(i.NoActivityDays)
	return nil
}

// Decode implements bin.Decoder.
func (i *InputBusinessGreetingMessage) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputBusinessGreetingMessage#194cb3b to nil")
	}
	if err := b.ConsumeID(InputBusinessGreetingMessageTypeID); err != nil {
		return fmt.Errorf("unable to decode inputBusinessGreetingMessage#194cb3b: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputBusinessGreetingMessage) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputBusinessGreetingMessage#194cb3b to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputBusinessGreetingMessage#194cb3b: field shortcut_id: %w", err)
		}
		i.ShortcutID = value
	}
	{
		if err := i.Recipients.Decode(b); err != nil {
			return fmt.Errorf("unable to decode inputBusinessGreetingMessage#194cb3b: field recipients: %w", err)
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputBusinessGreetingMessage#194cb3b: field no_activity_days: %w", err)
		}
		i.NoActivityDays = value
	}
	return nil
}

// GetShortcutID returns value of ShortcutID field.
func (i *InputBusinessGreetingMessage) GetShortcutID() (value int) {
	if i == nil {
		return
	}
	return i.ShortcutID
}

// GetRecipients returns value of Recipients field.
func (i *InputBusinessGreetingMessage) GetRecipients() (value InputBusinessRecipients) {
	if i == nil {
		return
	}
	return i.Recipients
}

// GetNoActivityDays returns value of NoActivityDays field.
func (i *InputBusinessGreetingMessage) GetNoActivityDays() (value int) {
	if i == nil {
		return
	}
	return i.NoActivityDays
}
