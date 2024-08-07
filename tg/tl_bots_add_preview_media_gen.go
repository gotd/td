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

// BotsAddPreviewMediaRequest represents TL type `bots.addPreviewMedia#17aeb75a`.
//
// See https://core.telegram.org/method/bots.addPreviewMedia for reference.
type BotsAddPreviewMediaRequest struct {
	// Bot field of BotsAddPreviewMediaRequest.
	Bot InputUserClass
	// LangCode field of BotsAddPreviewMediaRequest.
	LangCode string
	// Media field of BotsAddPreviewMediaRequest.
	Media InputMediaClass
}

// BotsAddPreviewMediaRequestTypeID is TL type id of BotsAddPreviewMediaRequest.
const BotsAddPreviewMediaRequestTypeID = 0x17aeb75a

// Ensuring interfaces in compile-time for BotsAddPreviewMediaRequest.
var (
	_ bin.Encoder     = &BotsAddPreviewMediaRequest{}
	_ bin.Decoder     = &BotsAddPreviewMediaRequest{}
	_ bin.BareEncoder = &BotsAddPreviewMediaRequest{}
	_ bin.BareDecoder = &BotsAddPreviewMediaRequest{}
)

func (a *BotsAddPreviewMediaRequest) Zero() bool {
	if a == nil {
		return true
	}
	if !(a.Bot == nil) {
		return false
	}
	if !(a.LangCode == "") {
		return false
	}
	if !(a.Media == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (a *BotsAddPreviewMediaRequest) String() string {
	if a == nil {
		return "BotsAddPreviewMediaRequest(nil)"
	}
	type Alias BotsAddPreviewMediaRequest
	return fmt.Sprintf("BotsAddPreviewMediaRequest%+v", Alias(*a))
}

// FillFrom fills BotsAddPreviewMediaRequest from given interface.
func (a *BotsAddPreviewMediaRequest) FillFrom(from interface {
	GetBot() (value InputUserClass)
	GetLangCode() (value string)
	GetMedia() (value InputMediaClass)
}) {
	a.Bot = from.GetBot()
	a.LangCode = from.GetLangCode()
	a.Media = from.GetMedia()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*BotsAddPreviewMediaRequest) TypeID() uint32 {
	return BotsAddPreviewMediaRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*BotsAddPreviewMediaRequest) TypeName() string {
	return "bots.addPreviewMedia"
}

// TypeInfo returns info about TL type.
func (a *BotsAddPreviewMediaRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "bots.addPreviewMedia",
		ID:   BotsAddPreviewMediaRequestTypeID,
	}
	if a == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Bot",
			SchemaName: "bot",
		},
		{
			Name:       "LangCode",
			SchemaName: "lang_code",
		},
		{
			Name:       "Media",
			SchemaName: "media",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (a *BotsAddPreviewMediaRequest) Encode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode bots.addPreviewMedia#17aeb75a as nil")
	}
	b.PutID(BotsAddPreviewMediaRequestTypeID)
	return a.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (a *BotsAddPreviewMediaRequest) EncodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode bots.addPreviewMedia#17aeb75a as nil")
	}
	if a.Bot == nil {
		return fmt.Errorf("unable to encode bots.addPreviewMedia#17aeb75a: field bot is nil")
	}
	if err := a.Bot.Encode(b); err != nil {
		return fmt.Errorf("unable to encode bots.addPreviewMedia#17aeb75a: field bot: %w", err)
	}
	b.PutString(a.LangCode)
	if a.Media == nil {
		return fmt.Errorf("unable to encode bots.addPreviewMedia#17aeb75a: field media is nil")
	}
	if err := a.Media.Encode(b); err != nil {
		return fmt.Errorf("unable to encode bots.addPreviewMedia#17aeb75a: field media: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (a *BotsAddPreviewMediaRequest) Decode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode bots.addPreviewMedia#17aeb75a to nil")
	}
	if err := b.ConsumeID(BotsAddPreviewMediaRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode bots.addPreviewMedia#17aeb75a: %w", err)
	}
	return a.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (a *BotsAddPreviewMediaRequest) DecodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode bots.addPreviewMedia#17aeb75a to nil")
	}
	{
		value, err := DecodeInputUser(b)
		if err != nil {
			return fmt.Errorf("unable to decode bots.addPreviewMedia#17aeb75a: field bot: %w", err)
		}
		a.Bot = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode bots.addPreviewMedia#17aeb75a: field lang_code: %w", err)
		}
		a.LangCode = value
	}
	{
		value, err := DecodeInputMedia(b)
		if err != nil {
			return fmt.Errorf("unable to decode bots.addPreviewMedia#17aeb75a: field media: %w", err)
		}
		a.Media = value
	}
	return nil
}

// GetBot returns value of Bot field.
func (a *BotsAddPreviewMediaRequest) GetBot() (value InputUserClass) {
	if a == nil {
		return
	}
	return a.Bot
}

// GetLangCode returns value of LangCode field.
func (a *BotsAddPreviewMediaRequest) GetLangCode() (value string) {
	if a == nil {
		return
	}
	return a.LangCode
}

// GetMedia returns value of Media field.
func (a *BotsAddPreviewMediaRequest) GetMedia() (value InputMediaClass) {
	if a == nil {
		return
	}
	return a.Media
}

// BotsAddPreviewMedia invokes method bots.addPreviewMedia#17aeb75a returning error if any.
//
// See https://core.telegram.org/method/bots.addPreviewMedia for reference.
func (c *Client) BotsAddPreviewMedia(ctx context.Context, request *BotsAddPreviewMediaRequest) (*BotPreviewMedia, error) {
	var result BotPreviewMedia

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
