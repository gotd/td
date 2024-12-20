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

// ChatAffiliateProgram represents TL type `chatAffiliateProgram#ab9c1136`.
type ChatAffiliateProgram struct {
	// The link that can be used to refer users if the program is still active
	URL string
	// User identifier of the bot created the program
	BotUserID int64
	// The parameters of the affiliate program
	Parameters AffiliateProgramParameters
	// Point in time (Unix timestamp) when the affiliate program was connected
	ConnectionDate int32
	// True, if the program was canceled by the bot, or disconnected by the chat owner and
	// isn't available anymore
	IsDisconnected bool
	// The number of users that used the affiliate program
	UserCount int64
	// The number of Telegram Stars that were earned by the affiliate program
	RevenueStarCount int64
}

// ChatAffiliateProgramTypeID is TL type id of ChatAffiliateProgram.
const ChatAffiliateProgramTypeID = 0xab9c1136

// Ensuring interfaces in compile-time for ChatAffiliateProgram.
var (
	_ bin.Encoder     = &ChatAffiliateProgram{}
	_ bin.Decoder     = &ChatAffiliateProgram{}
	_ bin.BareEncoder = &ChatAffiliateProgram{}
	_ bin.BareDecoder = &ChatAffiliateProgram{}
)

func (c *ChatAffiliateProgram) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.URL == "") {
		return false
	}
	if !(c.BotUserID == 0) {
		return false
	}
	if !(c.Parameters.Zero()) {
		return false
	}
	if !(c.ConnectionDate == 0) {
		return false
	}
	if !(c.IsDisconnected == false) {
		return false
	}
	if !(c.UserCount == 0) {
		return false
	}
	if !(c.RevenueStarCount == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *ChatAffiliateProgram) String() string {
	if c == nil {
		return "ChatAffiliateProgram(nil)"
	}
	type Alias ChatAffiliateProgram
	return fmt.Sprintf("ChatAffiliateProgram%+v", Alias(*c))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChatAffiliateProgram) TypeID() uint32 {
	return ChatAffiliateProgramTypeID
}

// TypeName returns name of type in TL schema.
func (*ChatAffiliateProgram) TypeName() string {
	return "chatAffiliateProgram"
}

// TypeInfo returns info about TL type.
func (c *ChatAffiliateProgram) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "chatAffiliateProgram",
		ID:   ChatAffiliateProgramTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "URL",
			SchemaName: "url",
		},
		{
			Name:       "BotUserID",
			SchemaName: "bot_user_id",
		},
		{
			Name:       "Parameters",
			SchemaName: "parameters",
		},
		{
			Name:       "ConnectionDate",
			SchemaName: "connection_date",
		},
		{
			Name:       "IsDisconnected",
			SchemaName: "is_disconnected",
		},
		{
			Name:       "UserCount",
			SchemaName: "user_count",
		},
		{
			Name:       "RevenueStarCount",
			SchemaName: "revenue_star_count",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *ChatAffiliateProgram) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode chatAffiliateProgram#ab9c1136 as nil")
	}
	b.PutID(ChatAffiliateProgramTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *ChatAffiliateProgram) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode chatAffiliateProgram#ab9c1136 as nil")
	}
	b.PutString(c.URL)
	b.PutInt53(c.BotUserID)
	if err := c.Parameters.Encode(b); err != nil {
		return fmt.Errorf("unable to encode chatAffiliateProgram#ab9c1136: field parameters: %w", err)
	}
	b.PutInt32(c.ConnectionDate)
	b.PutBool(c.IsDisconnected)
	b.PutLong(c.UserCount)
	b.PutLong(c.RevenueStarCount)
	return nil
}

// Decode implements bin.Decoder.
func (c *ChatAffiliateProgram) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode chatAffiliateProgram#ab9c1136 to nil")
	}
	if err := b.ConsumeID(ChatAffiliateProgramTypeID); err != nil {
		return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *ChatAffiliateProgram) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode chatAffiliateProgram#ab9c1136 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field url: %w", err)
		}
		c.URL = value
	}
	{
		value, err := b.Int53()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field bot_user_id: %w", err)
		}
		c.BotUserID = value
	}
	{
		if err := c.Parameters.Decode(b); err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field parameters: %w", err)
		}
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field connection_date: %w", err)
		}
		c.ConnectionDate = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field is_disconnected: %w", err)
		}
		c.IsDisconnected = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field user_count: %w", err)
		}
		c.UserCount = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field revenue_star_count: %w", err)
		}
		c.RevenueStarCount = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (c *ChatAffiliateProgram) EncodeTDLibJSON(b tdjson.Encoder) error {
	if c == nil {
		return fmt.Errorf("can't encode chatAffiliateProgram#ab9c1136 as nil")
	}
	b.ObjStart()
	b.PutID("chatAffiliateProgram")
	b.Comma()
	b.FieldStart("url")
	b.PutString(c.URL)
	b.Comma()
	b.FieldStart("bot_user_id")
	b.PutInt53(c.BotUserID)
	b.Comma()
	b.FieldStart("parameters")
	if err := c.Parameters.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode chatAffiliateProgram#ab9c1136: field parameters: %w", err)
	}
	b.Comma()
	b.FieldStart("connection_date")
	b.PutInt32(c.ConnectionDate)
	b.Comma()
	b.FieldStart("is_disconnected")
	b.PutBool(c.IsDisconnected)
	b.Comma()
	b.FieldStart("user_count")
	b.PutLong(c.UserCount)
	b.Comma()
	b.FieldStart("revenue_star_count")
	b.PutLong(c.RevenueStarCount)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (c *ChatAffiliateProgram) DecodeTDLibJSON(b tdjson.Decoder) error {
	if c == nil {
		return fmt.Errorf("can't decode chatAffiliateProgram#ab9c1136 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("chatAffiliateProgram"); err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: %w", err)
			}
		case "url":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field url: %w", err)
			}
			c.URL = value
		case "bot_user_id":
			value, err := b.Int53()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field bot_user_id: %w", err)
			}
			c.BotUserID = value
		case "parameters":
			if err := c.Parameters.DecodeTDLibJSON(b); err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field parameters: %w", err)
			}
		case "connection_date":
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field connection_date: %w", err)
			}
			c.ConnectionDate = value
		case "is_disconnected":
			value, err := b.Bool()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field is_disconnected: %w", err)
			}
			c.IsDisconnected = value
		case "user_count":
			value, err := b.Long()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field user_count: %w", err)
			}
			c.UserCount = value
		case "revenue_star_count":
			value, err := b.Long()
			if err != nil {
				return fmt.Errorf("unable to decode chatAffiliateProgram#ab9c1136: field revenue_star_count: %w", err)
			}
			c.RevenueStarCount = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetURL returns value of URL field.
func (c *ChatAffiliateProgram) GetURL() (value string) {
	if c == nil {
		return
	}
	return c.URL
}

// GetBotUserID returns value of BotUserID field.
func (c *ChatAffiliateProgram) GetBotUserID() (value int64) {
	if c == nil {
		return
	}
	return c.BotUserID
}

// GetParameters returns value of Parameters field.
func (c *ChatAffiliateProgram) GetParameters() (value AffiliateProgramParameters) {
	if c == nil {
		return
	}
	return c.Parameters
}

// GetConnectionDate returns value of ConnectionDate field.
func (c *ChatAffiliateProgram) GetConnectionDate() (value int32) {
	if c == nil {
		return
	}
	return c.ConnectionDate
}

// GetIsDisconnected returns value of IsDisconnected field.
func (c *ChatAffiliateProgram) GetIsDisconnected() (value bool) {
	if c == nil {
		return
	}
	return c.IsDisconnected
}

// GetUserCount returns value of UserCount field.
func (c *ChatAffiliateProgram) GetUserCount() (value int64) {
	if c == nil {
		return
	}
	return c.UserCount
}

// GetRevenueStarCount returns value of RevenueStarCount field.
func (c *ChatAffiliateProgram) GetRevenueStarCount() (value int64) {
	if c == nil {
		return
	}
	return c.RevenueStarCount
}
