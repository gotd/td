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

// MessagesAcceptURLAuthRequest represents TL type `messages.acceptUrlAuth#b12c7125`.
// Use this to accept a Seamless Telegram Login authorization request, for more info
// click here »¹
//
// Links:
//  1) https://core.telegram.org/api/url-authorization
//
// See https://core.telegram.org/method/messages.acceptUrlAuth for reference.
type MessagesAcceptURLAuthRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Set this flag to allow the bot to send messages to you (if requested)
	WriteAllowed bool
	// The location of the message
	//
	// Use SetPeer and GetPeer helpers.
	Peer InputPeerClass
	// Message ID of the message with the login button
	//
	// Use SetMsgID and GetMsgID helpers.
	MsgID int
	// ID of the login button
	//
	// Use SetButtonID and GetButtonID helpers.
	ButtonID int
	// URL field of MessagesAcceptURLAuthRequest.
	//
	// Use SetURL and GetURL helpers.
	URL string
}

// MessagesAcceptURLAuthRequestTypeID is TL type id of MessagesAcceptURLAuthRequest.
const MessagesAcceptURLAuthRequestTypeID = 0xb12c7125

func (a *MessagesAcceptURLAuthRequest) Zero() bool {
	if a == nil {
		return true
	}
	if !(a.Flags.Zero()) {
		return false
	}
	if !(a.WriteAllowed == false) {
		return false
	}
	if !(a.Peer == nil) {
		return false
	}
	if !(a.MsgID == 0) {
		return false
	}
	if !(a.ButtonID == 0) {
		return false
	}
	if !(a.URL == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (a *MessagesAcceptURLAuthRequest) String() string {
	if a == nil {
		return "MessagesAcceptURLAuthRequest(nil)"
	}
	type Alias MessagesAcceptURLAuthRequest
	return fmt.Sprintf("MessagesAcceptURLAuthRequest%+v", Alias(*a))
}

// FillFrom fills MessagesAcceptURLAuthRequest from given interface.
func (a *MessagesAcceptURLAuthRequest) FillFrom(from interface {
	GetWriteAllowed() (value bool)
	GetPeer() (value InputPeerClass, ok bool)
	GetMsgID() (value int, ok bool)
	GetButtonID() (value int, ok bool)
	GetURL() (value string, ok bool)
}) {
	a.WriteAllowed = from.GetWriteAllowed()
	if val, ok := from.GetPeer(); ok {
		a.Peer = val
	}

	if val, ok := from.GetMsgID(); ok {
		a.MsgID = val
	}

	if val, ok := from.GetButtonID(); ok {
		a.ButtonID = val
	}

	if val, ok := from.GetURL(); ok {
		a.URL = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesAcceptURLAuthRequest) TypeID() uint32 {
	return MessagesAcceptURLAuthRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesAcceptURLAuthRequest) TypeName() string {
	return "messages.acceptUrlAuth"
}

// TypeInfo returns info about TL type.
func (a *MessagesAcceptURLAuthRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.acceptUrlAuth",
		ID:   MessagesAcceptURLAuthRequestTypeID,
	}
	if a == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "WriteAllowed",
			SchemaName: "write_allowed",
			Null:       !a.Flags.Has(0),
		},
		{
			Name:       "Peer",
			SchemaName: "peer",
			Null:       !a.Flags.Has(1),
		},
		{
			Name:       "MsgID",
			SchemaName: "msg_id",
			Null:       !a.Flags.Has(1),
		},
		{
			Name:       "ButtonID",
			SchemaName: "button_id",
			Null:       !a.Flags.Has(1),
		},
		{
			Name:       "URL",
			SchemaName: "url",
			Null:       !a.Flags.Has(2),
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (a *MessagesAcceptURLAuthRequest) Encode(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.acceptUrlAuth#b12c7125",
		}
	}
	b.PutID(MessagesAcceptURLAuthRequestTypeID)
	return a.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (a *MessagesAcceptURLAuthRequest) EncodeBare(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.acceptUrlAuth#b12c7125",
		}
	}
	if !(a.WriteAllowed == false) {
		a.Flags.Set(0)
	}
	if !(a.Peer == nil) {
		a.Flags.Set(1)
	}
	if !(a.MsgID == 0) {
		a.Flags.Set(1)
	}
	if !(a.ButtonID == 0) {
		a.Flags.Set(1)
	}
	if !(a.URL == "") {
		a.Flags.Set(2)
	}
	if err := a.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.acceptUrlAuth#b12c7125",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	if a.Flags.Has(1) {
		if a.Peer == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.acceptUrlAuth#b12c7125",
				FieldName: "peer",
				Underlying: &bin.NilError{
					Action:   "encode",
					TypeName: "InputPeer",
				},
			}
		}
		if err := a.Peer.Encode(b); err != nil {
			return &bin.FieldError{
				Action:     "encode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "peer",
				Underlying: err,
			}
		}
	}
	if a.Flags.Has(1) {
		b.PutInt(a.MsgID)
	}
	if a.Flags.Has(1) {
		b.PutInt(a.ButtonID)
	}
	if a.Flags.Has(2) {
		b.PutString(a.URL)
	}
	return nil
}

// SetWriteAllowed sets value of WriteAllowed conditional field.
func (a *MessagesAcceptURLAuthRequest) SetWriteAllowed(value bool) {
	if value {
		a.Flags.Set(0)
		a.WriteAllowed = true
	} else {
		a.Flags.Unset(0)
		a.WriteAllowed = false
	}
}

// GetWriteAllowed returns value of WriteAllowed conditional field.
func (a *MessagesAcceptURLAuthRequest) GetWriteAllowed() (value bool) {
	return a.Flags.Has(0)
}

// SetPeer sets value of Peer conditional field.
func (a *MessagesAcceptURLAuthRequest) SetPeer(value InputPeerClass) {
	a.Flags.Set(1)
	a.Peer = value
}

// GetPeer returns value of Peer conditional field and
// boolean which is true if field was set.
func (a *MessagesAcceptURLAuthRequest) GetPeer() (value InputPeerClass, ok bool) {
	if !a.Flags.Has(1) {
		return value, false
	}
	return a.Peer, true
}

// SetMsgID sets value of MsgID conditional field.
func (a *MessagesAcceptURLAuthRequest) SetMsgID(value int) {
	a.Flags.Set(1)
	a.MsgID = value
}

// GetMsgID returns value of MsgID conditional field and
// boolean which is true if field was set.
func (a *MessagesAcceptURLAuthRequest) GetMsgID() (value int, ok bool) {
	if !a.Flags.Has(1) {
		return value, false
	}
	return a.MsgID, true
}

// SetButtonID sets value of ButtonID conditional field.
func (a *MessagesAcceptURLAuthRequest) SetButtonID(value int) {
	a.Flags.Set(1)
	a.ButtonID = value
}

// GetButtonID returns value of ButtonID conditional field and
// boolean which is true if field was set.
func (a *MessagesAcceptURLAuthRequest) GetButtonID() (value int, ok bool) {
	if !a.Flags.Has(1) {
		return value, false
	}
	return a.ButtonID, true
}

// SetURL sets value of URL conditional field.
func (a *MessagesAcceptURLAuthRequest) SetURL(value string) {
	a.Flags.Set(2)
	a.URL = value
}

// GetURL returns value of URL conditional field and
// boolean which is true if field was set.
func (a *MessagesAcceptURLAuthRequest) GetURL() (value string, ok bool) {
	if !a.Flags.Has(2) {
		return value, false
	}
	return a.URL, true
}

// Decode implements bin.Decoder.
func (a *MessagesAcceptURLAuthRequest) Decode(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.acceptUrlAuth#b12c7125",
		}
	}
	if err := b.ConsumeID(MessagesAcceptURLAuthRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.acceptUrlAuth#b12c7125",
			Underlying: err,
		}
	}
	return a.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (a *MessagesAcceptURLAuthRequest) DecodeBare(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.acceptUrlAuth#b12c7125",
		}
	}
	{
		if err := a.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	a.WriteAllowed = a.Flags.Has(0)
	if a.Flags.Has(1) {
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		a.Peer = value
	}
	if a.Flags.Has(1) {
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "msg_id",
				Underlying: err,
			}
		}
		a.MsgID = value
	}
	if a.Flags.Has(1) {
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "button_id",
				Underlying: err,
			}
		}
		a.ButtonID = value
	}
	if a.Flags.Has(2) {
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.acceptUrlAuth#b12c7125",
				FieldName:  "url",
				Underlying: err,
			}
		}
		a.URL = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesAcceptURLAuthRequest.
var (
	_ bin.Encoder     = &MessagesAcceptURLAuthRequest{}
	_ bin.Decoder     = &MessagesAcceptURLAuthRequest{}
	_ bin.BareEncoder = &MessagesAcceptURLAuthRequest{}
	_ bin.BareDecoder = &MessagesAcceptURLAuthRequest{}
)

// MessagesAcceptURLAuth invokes method messages.acceptUrlAuth#b12c7125 returning error if any.
// Use this to accept a Seamless Telegram Login authorization request, for more info
// click here »¹
//
// Links:
//  1) https://core.telegram.org/api/url-authorization
//
// See https://core.telegram.org/method/messages.acceptUrlAuth for reference.
func (c *Client) MessagesAcceptURLAuth(ctx context.Context, request *MessagesAcceptURLAuthRequest) (URLAuthResultClass, error) {
	var result URLAuthResultBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.UrlAuthResult, nil
}
