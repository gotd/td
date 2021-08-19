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

// StatsGetMegagroupStatsRequest represents TL type `stats.getMegagroupStats#dcdf8607`.
// Get supergroup statistics¹
//
// Links:
//  1) https://core.telegram.org/api/stats
//
// See https://core.telegram.org/method/stats.getMegagroupStats for reference.
type StatsGetMegagroupStatsRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Whether to enable dark theme for graph colors
	Dark bool
	// Supergroup ID¹
	//
	// Links:
	//  1) https://core.telegram.org/api/channel
	Channel InputChannelClass
}

// StatsGetMegagroupStatsRequestTypeID is TL type id of StatsGetMegagroupStatsRequest.
const StatsGetMegagroupStatsRequestTypeID = 0xdcdf8607

func (g *StatsGetMegagroupStatsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.Dark == false) {
		return false
	}
	if !(g.Channel == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *StatsGetMegagroupStatsRequest) String() string {
	if g == nil {
		return "StatsGetMegagroupStatsRequest(nil)"
	}
	type Alias StatsGetMegagroupStatsRequest
	return fmt.Sprintf("StatsGetMegagroupStatsRequest%+v", Alias(*g))
}

// FillFrom fills StatsGetMegagroupStatsRequest from given interface.
func (g *StatsGetMegagroupStatsRequest) FillFrom(from interface {
	GetDark() (value bool)
	GetChannel() (value InputChannelClass)
}) {
	g.Dark = from.GetDark()
	g.Channel = from.GetChannel()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*StatsGetMegagroupStatsRequest) TypeID() uint32 {
	return StatsGetMegagroupStatsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*StatsGetMegagroupStatsRequest) TypeName() string {
	return "stats.getMegagroupStats"
}

// TypeInfo returns info about TL type.
func (g *StatsGetMegagroupStatsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "stats.getMegagroupStats",
		ID:   StatsGetMegagroupStatsRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Dark",
			SchemaName: "dark",
			Null:       !g.Flags.Has(0),
		},
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *StatsGetMegagroupStatsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "stats.getMegagroupStats#dcdf8607",
		}
	}
	b.PutID(StatsGetMegagroupStatsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *StatsGetMegagroupStatsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "stats.getMegagroupStats#dcdf8607",
		}
	}
	if !(g.Dark == false) {
		g.Flags.Set(0)
	}
	if err := g.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "stats.getMegagroupStats#dcdf8607",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	if g.Channel == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "stats.getMegagroupStats#dcdf8607",
			FieldName: "channel",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputChannel",
			},
		}
	}
	if err := g.Channel.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "stats.getMegagroupStats#dcdf8607",
			FieldName:  "channel",
			Underlying: err,
		}
	}
	return nil
}

// SetDark sets value of Dark conditional field.
func (g *StatsGetMegagroupStatsRequest) SetDark(value bool) {
	if value {
		g.Flags.Set(0)
		g.Dark = true
	} else {
		g.Flags.Unset(0)
		g.Dark = false
	}
}

// GetDark returns value of Dark conditional field.
func (g *StatsGetMegagroupStatsRequest) GetDark() (value bool) {
	return g.Flags.Has(0)
}

// GetChannel returns value of Channel field.
func (g *StatsGetMegagroupStatsRequest) GetChannel() (value InputChannelClass) {
	return g.Channel
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (g *StatsGetMegagroupStatsRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return g.Channel.AsNotEmpty()
}

// Decode implements bin.Decoder.
func (g *StatsGetMegagroupStatsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "stats.getMegagroupStats#dcdf8607",
		}
	}
	if err := b.ConsumeID(StatsGetMegagroupStatsRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "stats.getMegagroupStats#dcdf8607",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *StatsGetMegagroupStatsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "stats.getMegagroupStats#dcdf8607",
		}
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "stats.getMegagroupStats#dcdf8607",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	g.Dark = g.Flags.Has(0)
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "stats.getMegagroupStats#dcdf8607",
				FieldName:  "channel",
				Underlying: err,
			}
		}
		g.Channel = value
	}
	return nil
}

// Ensuring interfaces in compile-time for StatsGetMegagroupStatsRequest.
var (
	_ bin.Encoder     = &StatsGetMegagroupStatsRequest{}
	_ bin.Decoder     = &StatsGetMegagroupStatsRequest{}
	_ bin.BareEncoder = &StatsGetMegagroupStatsRequest{}
	_ bin.BareDecoder = &StatsGetMegagroupStatsRequest{}
)

// StatsGetMegagroupStats invokes method stats.getMegagroupStats#dcdf8607 returning error if any.
// Get supergroup statistics¹
//
// Links:
//  1) https://core.telegram.org/api/stats
//
// See https://core.telegram.org/method/stats.getMegagroupStats for reference.
// Can be used by bots.
func (c *Client) StatsGetMegagroupStats(ctx context.Context, request *StatsGetMegagroupStatsRequest) (*StatsMegagroupStats, error) {
	var result StatsMegagroupStats

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
