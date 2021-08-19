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

// HelpGetPromoDataRequest represents TL type `help.getPromoData#c0977421`.
// Get MTProxy/Public Service Announcement information
//
// See https://core.telegram.org/method/help.getPromoData for reference.
type HelpGetPromoDataRequest struct {
}

// HelpGetPromoDataRequestTypeID is TL type id of HelpGetPromoDataRequest.
const HelpGetPromoDataRequestTypeID = 0xc0977421

func (g *HelpGetPromoDataRequest) Zero() bool {
	if g == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (g *HelpGetPromoDataRequest) String() string {
	if g == nil {
		return "HelpGetPromoDataRequest(nil)"
	}
	type Alias HelpGetPromoDataRequest
	return fmt.Sprintf("HelpGetPromoDataRequest%+v", Alias(*g))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpGetPromoDataRequest) TypeID() uint32 {
	return HelpGetPromoDataRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpGetPromoDataRequest) TypeName() string {
	return "help.getPromoData"
}

// TypeInfo returns info about TL type.
func (g *HelpGetPromoDataRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.getPromoData",
		ID:   HelpGetPromoDataRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (g *HelpGetPromoDataRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.getPromoData#c0977421",
		}
	}
	b.PutID(HelpGetPromoDataRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *HelpGetPromoDataRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "help.getPromoData#c0977421",
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *HelpGetPromoDataRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.getPromoData#c0977421",
		}
	}
	if err := b.ConsumeID(HelpGetPromoDataRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "help.getPromoData#c0977421",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *HelpGetPromoDataRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "help.getPromoData#c0977421",
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for HelpGetPromoDataRequest.
var (
	_ bin.Encoder     = &HelpGetPromoDataRequest{}
	_ bin.Decoder     = &HelpGetPromoDataRequest{}
	_ bin.BareEncoder = &HelpGetPromoDataRequest{}
	_ bin.BareDecoder = &HelpGetPromoDataRequest{}
)

// HelpGetPromoData invokes method help.getPromoData#c0977421 returning error if any.
// Get MTProxy/Public Service Announcement information
//
// See https://core.telegram.org/method/help.getPromoData for reference.
// Can be used by bots.
func (c *Client) HelpGetPromoData(ctx context.Context) (HelpPromoDataClass, error) {
	var result HelpPromoDataBox

	request := &HelpGetPromoDataRequest{}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.PromoData, nil
}
