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

// MessagesGetSearchCountersRequest represents TL type `messages.getSearchCounters#732eef00`.
// Get the number of results that would be found by a messages.search¹ call with the
// same parameters
//
// Links:
//  1) https://core.telegram.org/method/messages.search
//
// See https://core.telegram.org/method/messages.getSearchCounters for reference.
type MessagesGetSearchCountersRequest struct {
	// Peer where to search
	Peer InputPeerClass
	// Search filters
	Filters []MessagesFilterClass
}

// MessagesGetSearchCountersRequestTypeID is TL type id of MessagesGetSearchCountersRequest.
const MessagesGetSearchCountersRequestTypeID = 0x732eef00

func (g *MessagesGetSearchCountersRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Peer == nil) {
		return false
	}
	if !(g.Filters == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *MessagesGetSearchCountersRequest) String() string {
	if g == nil {
		return "MessagesGetSearchCountersRequest(nil)"
	}
	type Alias MessagesGetSearchCountersRequest
	return fmt.Sprintf("MessagesGetSearchCountersRequest%+v", Alias(*g))
}

// FillFrom fills MessagesGetSearchCountersRequest from given interface.
func (g *MessagesGetSearchCountersRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
	GetFilters() (value []MessagesFilterClass)
}) {
	g.Peer = from.GetPeer()
	g.Filters = from.GetFilters()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesGetSearchCountersRequest) TypeID() uint32 {
	return MessagesGetSearchCountersRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesGetSearchCountersRequest) TypeName() string {
	return "messages.getSearchCounters"
}

// TypeInfo returns info about TL type.
func (g *MessagesGetSearchCountersRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.getSearchCounters",
		ID:   MessagesGetSearchCountersRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "Filters",
			SchemaName: "filters",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *MessagesGetSearchCountersRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getSearchCounters#732eef00",
		}
	}
	b.PutID(MessagesGetSearchCountersRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *MessagesGetSearchCountersRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.getSearchCounters#732eef00",
		}
	}
	if g.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "messages.getSearchCounters#732eef00",
			FieldName: "peer",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputPeer",
			},
		}
	}
	if err := g.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.getSearchCounters#732eef00",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	b.PutVectorHeader(len(g.Filters))
	for idx, v := range g.Filters {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.getSearchCounters#732eef00",
				FieldName: "filters",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<MessagesFilter>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.getSearchCounters#732eef00",
				FieldName: "filters",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *MessagesGetSearchCountersRequest) GetPeer() (value InputPeerClass) {
	return g.Peer
}

// GetFilters returns value of Filters field.
func (g *MessagesGetSearchCountersRequest) GetFilters() (value []MessagesFilterClass) {
	return g.Filters
}

// MapFilters returns field Filters wrapped in MessagesFilterClassArray helper.
func (g *MessagesGetSearchCountersRequest) MapFilters() (value MessagesFilterClassArray) {
	return MessagesFilterClassArray(g.Filters)
}

// Decode implements bin.Decoder.
func (g *MessagesGetSearchCountersRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getSearchCounters#732eef00",
		}
	}
	if err := b.ConsumeID(MessagesGetSearchCountersRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.getSearchCounters#732eef00",
			Underlying: err,
		}
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *MessagesGetSearchCountersRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.getSearchCounters#732eef00",
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.getSearchCounters#732eef00",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		g.Peer = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.getSearchCounters#732eef00",
				FieldName:  "filters",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			g.Filters = make([]MessagesFilterClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeMessagesFilter(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "messages.getSearchCounters#732eef00",
					FieldName:  "filters",
					Underlying: err,
				}
			}
			g.Filters = append(g.Filters, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesGetSearchCountersRequest.
var (
	_ bin.Encoder     = &MessagesGetSearchCountersRequest{}
	_ bin.Decoder     = &MessagesGetSearchCountersRequest{}
	_ bin.BareEncoder = &MessagesGetSearchCountersRequest{}
	_ bin.BareDecoder = &MessagesGetSearchCountersRequest{}
)

// MessagesGetSearchCounters invokes method messages.getSearchCounters#732eef00 returning error if any.
// Get the number of results that would be found by a messages.search¹ call with the
// same parameters
//
// Links:
//  1) https://core.telegram.org/method/messages.search
//
// See https://core.telegram.org/method/messages.getSearchCounters for reference.
func (c *Client) MessagesGetSearchCounters(ctx context.Context, request *MessagesGetSearchCountersRequest) ([]MessagesSearchCounter, error) {
	var result MessagesSearchCounterVector

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return []MessagesSearchCounter(result.Elems), nil
}
