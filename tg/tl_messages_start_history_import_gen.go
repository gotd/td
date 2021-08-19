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

// MessagesStartHistoryImportRequest represents TL type `messages.startHistoryImport#b43df344`.
//
// See https://core.telegram.org/method/messages.startHistoryImport for reference.
type MessagesStartHistoryImportRequest struct {
	// Peer field of MessagesStartHistoryImportRequest.
	Peer InputPeerClass
	// ImportID field of MessagesStartHistoryImportRequest.
	ImportID int64
}

// MessagesStartHistoryImportRequestTypeID is TL type id of MessagesStartHistoryImportRequest.
const MessagesStartHistoryImportRequestTypeID = 0xb43df344

func (s *MessagesStartHistoryImportRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Peer == nil) {
		return false
	}
	if !(s.ImportID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *MessagesStartHistoryImportRequest) String() string {
	if s == nil {
		return "MessagesStartHistoryImportRequest(nil)"
	}
	type Alias MessagesStartHistoryImportRequest
	return fmt.Sprintf("MessagesStartHistoryImportRequest%+v", Alias(*s))
}

// FillFrom fills MessagesStartHistoryImportRequest from given interface.
func (s *MessagesStartHistoryImportRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
	GetImportID() (value int64)
}) {
	s.Peer = from.GetPeer()
	s.ImportID = from.GetImportID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesStartHistoryImportRequest) TypeID() uint32 {
	return MessagesStartHistoryImportRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesStartHistoryImportRequest) TypeName() string {
	return "messages.startHistoryImport"
}

// TypeInfo returns info about TL type.
func (s *MessagesStartHistoryImportRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.startHistoryImport",
		ID:   MessagesStartHistoryImportRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "ImportID",
			SchemaName: "import_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *MessagesStartHistoryImportRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.startHistoryImport#b43df344",
		}
	}
	b.PutID(MessagesStartHistoryImportRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *MessagesStartHistoryImportRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.startHistoryImport#b43df344",
		}
	}
	if s.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "messages.startHistoryImport#b43df344",
			FieldName: "peer",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputPeer",
			},
		}
	}
	if err := s.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.startHistoryImport#b43df344",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	b.PutLong(s.ImportID)
	return nil
}

// GetPeer returns value of Peer field.
func (s *MessagesStartHistoryImportRequest) GetPeer() (value InputPeerClass) {
	return s.Peer
}

// GetImportID returns value of ImportID field.
func (s *MessagesStartHistoryImportRequest) GetImportID() (value int64) {
	return s.ImportID
}

// Decode implements bin.Decoder.
func (s *MessagesStartHistoryImportRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.startHistoryImport#b43df344",
		}
	}
	if err := b.ConsumeID(MessagesStartHistoryImportRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.startHistoryImport#b43df344",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *MessagesStartHistoryImportRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.startHistoryImport#b43df344",
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.startHistoryImport#b43df344",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		s.Peer = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.startHistoryImport#b43df344",
				FieldName:  "import_id",
				Underlying: err,
			}
		}
		s.ImportID = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesStartHistoryImportRequest.
var (
	_ bin.Encoder     = &MessagesStartHistoryImportRequest{}
	_ bin.Decoder     = &MessagesStartHistoryImportRequest{}
	_ bin.BareEncoder = &MessagesStartHistoryImportRequest{}
	_ bin.BareDecoder = &MessagesStartHistoryImportRequest{}
)

// MessagesStartHistoryImport invokes method messages.startHistoryImport#b43df344 returning error if any.
//
// See https://core.telegram.org/method/messages.startHistoryImport for reference.
func (c *Client) MessagesStartHistoryImport(ctx context.Context, request *MessagesStartHistoryImportRequest) (bool, error) {
	var result BoolBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
