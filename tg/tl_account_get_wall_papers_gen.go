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

// AccountGetWallPapersRequest represents TL type `account.getWallPapers#aabb1763`.
// Returns a list of available wallpapers.
//
// See https://core.telegram.org/method/account.getWallPapers for reference.
type AccountGetWallPapersRequest struct {
	// Hash for pagination, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/offsets#hash-generation
	Hash int
}

// AccountGetWallPapersRequestTypeID is TL type id of AccountGetWallPapersRequest.
const AccountGetWallPapersRequestTypeID = 0xaabb1763

func (g *AccountGetWallPapersRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Hash == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *AccountGetWallPapersRequest) String() string {
	if g == nil {
		return "AccountGetWallPapersRequest(nil)"
	}
	type Alias AccountGetWallPapersRequest
	return fmt.Sprintf("AccountGetWallPapersRequest%+v", Alias(*g))
}

// FillFrom fills AccountGetWallPapersRequest from given interface.
func (g *AccountGetWallPapersRequest) FillFrom(from interface {
	GetHash() (value int)
}) {
	g.Hash = from.GetHash()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountGetWallPapersRequest) TypeID() uint32 {
	return AccountGetWallPapersRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountGetWallPapersRequest) TypeName() string {
	return "account.getWallPapers"
}

// TypeInfo returns info about TL type.
func (g *AccountGetWallPapersRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.getWallPapers",
		ID:   AccountGetWallPapersRequestTypeID,
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
func (g *AccountGetWallPapersRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.getWallPapers#aabb1763",
		}
	}
	b.PutID(AccountGetWallPapersRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *AccountGetWallPapersRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.getWallPapers#aabb1763",
		}
	}
	b.PutInt(g.Hash)
	return nil
}

// GetHash returns value of Hash field.
func (g *AccountGetWallPapersRequest) GetHash() (value int) {
	return g.Hash
}

// Decode implements bin.Decoder.
func (g *AccountGetWallPapersRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.getWallPapers#aabb1763",
		}
	}
	if err := b.ConsumeID(AccountGetWallPapersRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "account.getWallPapers#aabb1763",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *AccountGetWallPapersRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.getWallPapers#aabb1763",
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "account.getWallPapers#aabb1763",
				FieldName:  "hash",
				Underlying: err,
			}
		}
		g.Hash = value
	}
	return nil
}

// Ensuring interfaces in compile-time for AccountGetWallPapersRequest.
var (
	_ bin.Encoder     = &AccountGetWallPapersRequest{}
	_ bin.Decoder     = &AccountGetWallPapersRequest{}
	_ bin.BareEncoder = &AccountGetWallPapersRequest{}
	_ bin.BareDecoder = &AccountGetWallPapersRequest{}
)

// AccountGetWallPapers invokes method account.getWallPapers#aabb1763 returning error if any.
// Returns a list of available wallpapers.
//
// See https://core.telegram.org/method/account.getWallPapers for reference.
func (c *Client) AccountGetWallPapers(ctx context.Context, hash int) (AccountWallPapersClass, error) {
	var result AccountWallPapersBox

	request := &AccountGetWallPapersRequest{
		Hash: hash,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.WallPapers, nil
}
