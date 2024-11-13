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

// MessagesClickSponsoredMessageRequest represents TL type `messages.clickSponsoredMessage#f093465`.
//
// See https://core.telegram.org/method/messages.clickSponsoredMessage for reference.
type MessagesClickSponsoredMessageRequest struct {
	// Flags field of MessagesClickSponsoredMessageRequest.
	Flags bin.Fields
	// Media field of MessagesClickSponsoredMessageRequest.
	Media bool
	// Fullscreen field of MessagesClickSponsoredMessageRequest.
	Fullscreen bool
	// Peer field of MessagesClickSponsoredMessageRequest.
	Peer InputPeerClass
	// RandomID field of MessagesClickSponsoredMessageRequest.
	RandomID []byte
}

// MessagesClickSponsoredMessageRequestTypeID is TL type id of MessagesClickSponsoredMessageRequest.
const MessagesClickSponsoredMessageRequestTypeID = 0xf093465

// Ensuring interfaces in compile-time for MessagesClickSponsoredMessageRequest.
var (
	_ bin.Encoder     = &MessagesClickSponsoredMessageRequest{}
	_ bin.Decoder     = &MessagesClickSponsoredMessageRequest{}
	_ bin.BareEncoder = &MessagesClickSponsoredMessageRequest{}
	_ bin.BareDecoder = &MessagesClickSponsoredMessageRequest{}
)

func (c *MessagesClickSponsoredMessageRequest) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.Flags.Zero()) {
		return false
	}
	if !(c.Media == false) {
		return false
	}
	if !(c.Fullscreen == false) {
		return false
	}
	if !(c.Peer == nil) {
		return false
	}
	if !(c.RandomID == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *MessagesClickSponsoredMessageRequest) String() string {
	if c == nil {
		return "MessagesClickSponsoredMessageRequest(nil)"
	}
	type Alias MessagesClickSponsoredMessageRequest
	return fmt.Sprintf("MessagesClickSponsoredMessageRequest%+v", Alias(*c))
}

// FillFrom fills MessagesClickSponsoredMessageRequest from given interface.
func (c *MessagesClickSponsoredMessageRequest) FillFrom(from interface {
	GetMedia() (value bool)
	GetFullscreen() (value bool)
	GetPeer() (value InputPeerClass)
	GetRandomID() (value []byte)
}) {
	c.Media = from.GetMedia()
	c.Fullscreen = from.GetFullscreen()
	c.Peer = from.GetPeer()
	c.RandomID = from.GetRandomID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesClickSponsoredMessageRequest) TypeID() uint32 {
	return MessagesClickSponsoredMessageRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesClickSponsoredMessageRequest) TypeName() string {
	return "messages.clickSponsoredMessage"
}

// TypeInfo returns info about TL type.
func (c *MessagesClickSponsoredMessageRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.clickSponsoredMessage",
		ID:   MessagesClickSponsoredMessageRequestTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Media",
			SchemaName: "media",
			Null:       !c.Flags.Has(0),
		},
		{
			Name:       "Fullscreen",
			SchemaName: "fullscreen",
			Null:       !c.Flags.Has(1),
		},
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "RandomID",
			SchemaName: "random_id",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (c *MessagesClickSponsoredMessageRequest) SetFlags() {
	if !(c.Media == false) {
		c.Flags.Set(0)
	}
	if !(c.Fullscreen == false) {
		c.Flags.Set(1)
	}
}

// Encode implements bin.Encoder.
func (c *MessagesClickSponsoredMessageRequest) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode messages.clickSponsoredMessage#f093465 as nil")
	}
	b.PutID(MessagesClickSponsoredMessageRequestTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *MessagesClickSponsoredMessageRequest) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode messages.clickSponsoredMessage#f093465 as nil")
	}
	c.SetFlags()
	if err := c.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.clickSponsoredMessage#f093465: field flags: %w", err)
	}
	if c.Peer == nil {
		return fmt.Errorf("unable to encode messages.clickSponsoredMessage#f093465: field peer is nil")
	}
	if err := c.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.clickSponsoredMessage#f093465: field peer: %w", err)
	}
	b.PutBytes(c.RandomID)
	return nil
}

// Decode implements bin.Decoder.
func (c *MessagesClickSponsoredMessageRequest) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode messages.clickSponsoredMessage#f093465 to nil")
	}
	if err := b.ConsumeID(MessagesClickSponsoredMessageRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.clickSponsoredMessage#f093465: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *MessagesClickSponsoredMessageRequest) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode messages.clickSponsoredMessage#f093465 to nil")
	}
	{
		if err := c.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode messages.clickSponsoredMessage#f093465: field flags: %w", err)
		}
	}
	c.Media = c.Flags.Has(0)
	c.Fullscreen = c.Flags.Has(1)
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.clickSponsoredMessage#f093465: field peer: %w", err)
		}
		c.Peer = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return fmt.Errorf("unable to decode messages.clickSponsoredMessage#f093465: field random_id: %w", err)
		}
		c.RandomID = value
	}
	return nil
}

// SetMedia sets value of Media conditional field.
func (c *MessagesClickSponsoredMessageRequest) SetMedia(value bool) {
	if value {
		c.Flags.Set(0)
		c.Media = true
	} else {
		c.Flags.Unset(0)
		c.Media = false
	}
}

// GetMedia returns value of Media conditional field.
func (c *MessagesClickSponsoredMessageRequest) GetMedia() (value bool) {
	if c == nil {
		return
	}
	return c.Flags.Has(0)
}

// SetFullscreen sets value of Fullscreen conditional field.
func (c *MessagesClickSponsoredMessageRequest) SetFullscreen(value bool) {
	if value {
		c.Flags.Set(1)
		c.Fullscreen = true
	} else {
		c.Flags.Unset(1)
		c.Fullscreen = false
	}
}

// GetFullscreen returns value of Fullscreen conditional field.
func (c *MessagesClickSponsoredMessageRequest) GetFullscreen() (value bool) {
	if c == nil {
		return
	}
	return c.Flags.Has(1)
}

// GetPeer returns value of Peer field.
func (c *MessagesClickSponsoredMessageRequest) GetPeer() (value InputPeerClass) {
	if c == nil {
		return
	}
	return c.Peer
}

// GetRandomID returns value of RandomID field.
func (c *MessagesClickSponsoredMessageRequest) GetRandomID() (value []byte) {
	if c == nil {
		return
	}
	return c.RandomID
}

// MessagesClickSponsoredMessage invokes method messages.clickSponsoredMessage#f093465 returning error if any.
//
// See https://core.telegram.org/method/messages.clickSponsoredMessage for reference.
func (c *Client) MessagesClickSponsoredMessage(ctx context.Context, request *MessagesClickSponsoredMessageRequest) (bool, error) {
	var result BoolBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}