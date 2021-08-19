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

// MessagesGetAllStickersRequest represents TL type `messages.getAllStickers#1c9618b1`.
// Get all installed stickers
//
// See https://core.telegram.org/method/messages.getAllStickers for reference.
type MessagesGetAllStickersRequest struct {
	// Hash for pagination, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets#hash-generation
	Hash int
}

// MessagesGetAllStickersRequestTypeID is TL type id of MessagesGetAllStickersRequest.
const MessagesGetAllStickersRequestTypeID = 0x1c9618b1

func (g *MessagesGetAllStickersRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Hash == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *MessagesGetAllStickersRequest) String() string {
	if g == nil {
		return "MessagesGetAllStickersRequest(nil)"
	}
	type Alias MessagesGetAllStickersRequest
	return fmt.Sprintf("MessagesGetAllStickersRequest%+v", Alias(*g))
}

// FillFrom fills MessagesGetAllStickersRequest from given interface.
func (g *MessagesGetAllStickersRequest) FillFrom(from interface {
	GetHash() (value int)
}) {
	g.Hash = from.GetHash()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesGetAllStickersRequest) TypeID() uint32 {
	return MessagesGetAllStickersRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesGetAllStickersRequest) TypeName() string {
	return "messages.getAllStickers"
}

// TypeInfo returns info about TL type.
func (g *MessagesGetAllStickersRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.getAllStickers",
		ID:   MessagesGetAllStickersRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Hash",
			SchemaName: "hash",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *MessagesGetAllStickersRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getAllStickers#1c9618b1",
		}
	}
	b.PutID(MessagesGetAllStickersRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *MessagesGetAllStickersRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getAllStickers#1c9618b1",
		}
	}
	b.PutInt(g.Hash)
	return nil
}

// GetHash returns value of Hash field.
func (g *MessagesGetAllStickersRequest) GetHash() (value int) {
	return g.Hash
}

// Decode implements bin.Decoder.
func (g *MessagesGetAllStickersRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getAllStickers#1c9618b1",
		}
	}
	if err := b.ConsumeID(MessagesGetAllStickersRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.getAllStickers#1c9618b1",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *MessagesGetAllStickersRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getAllStickers#1c9618b1",
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.getAllStickers#1c9618b1",
				FieldName:  "hash",
				Underlying: err,
			}
		}
		g.Hash = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesGetAllStickersRequest.
var (
	_ bin.Encoder     = &MessagesGetAllStickersRequest{}
	_ bin.Decoder     = &MessagesGetAllStickersRequest{}
	_ bin.BareEncoder = &MessagesGetAllStickersRequest{}
	_ bin.BareDecoder = &MessagesGetAllStickersRequest{}
)

// MessagesGetAllStickers invokes method messages.getAllStickers#1c9618b1 returning error if any.
// Get all installed stickers
//
// See https://core.telegram.org/method/messages.getAllStickers for reference.
func (c *Client) MessagesGetAllStickers(ctx context.Context, hash int) (MessagesAllStickersClass, error) {
	var result MessagesAllStickersBox

	request := &MessagesGetAllStickersRequest{
		Hash: hash,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.AllStickers, nil
}
