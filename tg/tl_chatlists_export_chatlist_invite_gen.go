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

// ChatlistsExportChatlistInviteRequest represents TL type `chatlists.exportChatlistInvite#8472478e`.
//
// See https://core.telegram.org/method/chatlists.exportChatlistInvite for reference.
type ChatlistsExportChatlistInviteRequest struct {
	// Chatlist field of ChatlistsExportChatlistInviteRequest.
	Chatlist InputChatlistDialogFilter
	// Title field of ChatlistsExportChatlistInviteRequest.
	Title string
	// Peers field of ChatlistsExportChatlistInviteRequest.
	Peers []InputPeerClass
}

// ChatlistsExportChatlistInviteRequestTypeID is TL type id of ChatlistsExportChatlistInviteRequest.
const ChatlistsExportChatlistInviteRequestTypeID = 0x8472478e

// Ensuring interfaces in compile-time for ChatlistsExportChatlistInviteRequest.
var (
	_ bin.Encoder     = &ChatlistsExportChatlistInviteRequest{}
	_ bin.Decoder     = &ChatlistsExportChatlistInviteRequest{}
	_ bin.BareEncoder = &ChatlistsExportChatlistInviteRequest{}
	_ bin.BareDecoder = &ChatlistsExportChatlistInviteRequest{}
)

func (e *ChatlistsExportChatlistInviteRequest) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Chatlist.Zero()) {
		return false
	}
	if !(e.Title == "") {
		return false
	}
	if !(e.Peers == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *ChatlistsExportChatlistInviteRequest) String() string {
	if e == nil {
		return "ChatlistsExportChatlistInviteRequest(nil)"
	}
	type Alias ChatlistsExportChatlistInviteRequest
	return fmt.Sprintf("ChatlistsExportChatlistInviteRequest%+v", Alias(*e))
}

// FillFrom fills ChatlistsExportChatlistInviteRequest from given interface.
func (e *ChatlistsExportChatlistInviteRequest) FillFrom(from interface {
	GetChatlist() (value InputChatlistDialogFilter)
	GetTitle() (value string)
	GetPeers() (value []InputPeerClass)
}) {
	e.Chatlist = from.GetChatlist()
	e.Title = from.GetTitle()
	e.Peers = from.GetPeers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChatlistsExportChatlistInviteRequest) TypeID() uint32 {
	return ChatlistsExportChatlistInviteRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ChatlistsExportChatlistInviteRequest) TypeName() string {
	return "chatlists.exportChatlistInvite"
}

// TypeInfo returns info about TL type.
func (e *ChatlistsExportChatlistInviteRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "chatlists.exportChatlistInvite",
		ID:   ChatlistsExportChatlistInviteRequestTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Chatlist",
			SchemaName: "chatlist",
		},
		{
			Name:       "Title",
			SchemaName: "title",
		},
		{
			Name:       "Peers",
			SchemaName: "peers",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (e *ChatlistsExportChatlistInviteRequest) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode chatlists.exportChatlistInvite#8472478e as nil")
	}
	b.PutID(ChatlistsExportChatlistInviteRequestTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *ChatlistsExportChatlistInviteRequest) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode chatlists.exportChatlistInvite#8472478e as nil")
	}
	if err := e.Chatlist.Encode(b); err != nil {
		return fmt.Errorf("unable to encode chatlists.exportChatlistInvite#8472478e: field chatlist: %w", err)
	}
	b.PutString(e.Title)
	b.PutVectorHeader(len(e.Peers))
	for idx, v := range e.Peers {
		if v == nil {
			return fmt.Errorf("unable to encode chatlists.exportChatlistInvite#8472478e: field peers element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode chatlists.exportChatlistInvite#8472478e: field peers element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *ChatlistsExportChatlistInviteRequest) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode chatlists.exportChatlistInvite#8472478e to nil")
	}
	if err := b.ConsumeID(ChatlistsExportChatlistInviteRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode chatlists.exportChatlistInvite#8472478e: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *ChatlistsExportChatlistInviteRequest) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode chatlists.exportChatlistInvite#8472478e to nil")
	}
	{
		if err := e.Chatlist.Decode(b); err != nil {
			return fmt.Errorf("unable to decode chatlists.exportChatlistInvite#8472478e: field chatlist: %w", err)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode chatlists.exportChatlistInvite#8472478e: field title: %w", err)
		}
		e.Title = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode chatlists.exportChatlistInvite#8472478e: field peers: %w", err)
		}

		if headerLen > 0 {
			e.Peers = make([]InputPeerClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputPeer(b)
			if err != nil {
				return fmt.Errorf("unable to decode chatlists.exportChatlistInvite#8472478e: field peers: %w", err)
			}
			e.Peers = append(e.Peers, value)
		}
	}
	return nil
}

// GetChatlist returns value of Chatlist field.
func (e *ChatlistsExportChatlistInviteRequest) GetChatlist() (value InputChatlistDialogFilter) {
	if e == nil {
		return
	}
	return e.Chatlist
}

// GetTitle returns value of Title field.
func (e *ChatlistsExportChatlistInviteRequest) GetTitle() (value string) {
	if e == nil {
		return
	}
	return e.Title
}

// GetPeers returns value of Peers field.
func (e *ChatlistsExportChatlistInviteRequest) GetPeers() (value []InputPeerClass) {
	if e == nil {
		return
	}
	return e.Peers
}

// MapPeers returns field Peers wrapped in InputPeerClassArray helper.
func (e *ChatlistsExportChatlistInviteRequest) MapPeers() (value InputPeerClassArray) {
	return InputPeerClassArray(e.Peers)
}

// ChatlistsExportChatlistInvite invokes method chatlists.exportChatlistInvite#8472478e returning error if any.
//
// Possible errors:
//
//	400 FILTER_ID_INVALID: The specified filter ID is invalid.
//
// See https://core.telegram.org/method/chatlists.exportChatlistInvite for reference.
func (c *Client) ChatlistsExportChatlistInvite(ctx context.Context, request *ChatlistsExportChatlistInviteRequest) (*ChatlistsExportedChatlistInvite, error) {
	var result ChatlistsExportedChatlistInvite

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
