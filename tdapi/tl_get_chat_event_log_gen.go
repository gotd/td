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

// GetChatEventLogRequest represents TL type `getChatEventLog#1e11b897`.
type GetChatEventLogRequest struct {
	// Chat identifier
	ChatID int64
	// Search query by which to filter events
	Query string
	// Identifier of an event from which to return results. Use 0 to get results from the
	// latest events
	FromEventID Int64
	// The maximum number of events to return; up to 100
	Limit int32
	// The types of events to return. By default, all types will be returned
	Filters ChatEventLogFilters
	// User identifiers by which to filter events. By default, events relating to all users
	// will be returned
	UserIDs []int32
}

// GetChatEventLogRequestTypeID is TL type id of GetChatEventLogRequest.
const GetChatEventLogRequestTypeID = 0x1e11b897

// Ensuring interfaces in compile-time for GetChatEventLogRequest.
var (
	_ bin.Encoder     = &GetChatEventLogRequest{}
	_ bin.Decoder     = &GetChatEventLogRequest{}
	_ bin.BareEncoder = &GetChatEventLogRequest{}
	_ bin.BareDecoder = &GetChatEventLogRequest{}
)

func (g *GetChatEventLogRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.ChatID == 0) {
		return false
	}
	if !(g.Query == "") {
		return false
	}
	if !(g.FromEventID.Zero()) {
		return false
	}
	if !(g.Limit == 0) {
		return false
	}
	if !(g.Filters.Zero()) {
		return false
	}
	if !(g.UserIDs == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *GetChatEventLogRequest) String() string {
	if g == nil {
		return "GetChatEventLogRequest(nil)"
	}
	type Alias GetChatEventLogRequest
	return fmt.Sprintf("GetChatEventLogRequest%+v", Alias(*g))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*GetChatEventLogRequest) TypeID() uint32 {
	return GetChatEventLogRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*GetChatEventLogRequest) TypeName() string {
	return "getChatEventLog"
}

// TypeInfo returns info about TL type.
func (g *GetChatEventLogRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "getChatEventLog",
		ID:   GetChatEventLogRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ChatID",
			SchemaName: "chat_id",
		},
		{
			Name:       "Query",
			SchemaName: "query",
		},
		{
			Name:       "FromEventID",
			SchemaName: "from_event_id",
		},
		{
			Name:       "Limit",
			SchemaName: "limit",
		},
		{
			Name:       "Filters",
			SchemaName: "filters",
		},
		{
			Name:       "UserIDs",
			SchemaName: "user_ids",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *GetChatEventLogRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode getChatEventLog#1e11b897 as nil")
	}
	b.PutID(GetChatEventLogRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *GetChatEventLogRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode getChatEventLog#1e11b897 as nil")
	}
	b.PutLong(g.ChatID)
	b.PutString(g.Query)
	if err := g.FromEventID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode getChatEventLog#1e11b897: field from_event_id: %w", err)
	}
	b.PutInt32(g.Limit)
	if err := g.Filters.Encode(b); err != nil {
		return fmt.Errorf("unable to encode getChatEventLog#1e11b897: field filters: %w", err)
	}
	b.PutInt(len(g.UserIDs))
	for _, v := range g.UserIDs {
		b.PutInt32(v)
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *GetChatEventLogRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode getChatEventLog#1e11b897 to nil")
	}
	if err := b.ConsumeID(GetChatEventLogRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode getChatEventLog#1e11b897: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *GetChatEventLogRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode getChatEventLog#1e11b897 to nil")
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field chat_id: %w", err)
		}
		g.ChatID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field query: %w", err)
		}
		g.Query = value
	}
	{
		if err := g.FromEventID.Decode(b); err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field from_event_id: %w", err)
		}
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field limit: %w", err)
		}
		g.Limit = value
	}
	{
		if err := g.Filters.Decode(b); err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field filters: %w", err)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field user_ids: %w", err)
		}

		if headerLen > 0 {
			g.UserIDs = make([]int32, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.Int32()
			if err != nil {
				return fmt.Errorf("unable to decode getChatEventLog#1e11b897: field user_ids: %w", err)
			}
			g.UserIDs = append(g.UserIDs, value)
		}
	}
	return nil
}

// GetChatID returns value of ChatID field.
func (g *GetChatEventLogRequest) GetChatID() (value int64) {
	return g.ChatID
}

// GetQuery returns value of Query field.
func (g *GetChatEventLogRequest) GetQuery() (value string) {
	return g.Query
}

// GetFromEventID returns value of FromEventID field.
func (g *GetChatEventLogRequest) GetFromEventID() (value Int64) {
	return g.FromEventID
}

// GetLimit returns value of Limit field.
func (g *GetChatEventLogRequest) GetLimit() (value int32) {
	return g.Limit
}

// GetFilters returns value of Filters field.
func (g *GetChatEventLogRequest) GetFilters() (value ChatEventLogFilters) {
	return g.Filters
}

// GetUserIDs returns value of UserIDs field.
func (g *GetChatEventLogRequest) GetUserIDs() (value []int32) {
	return g.UserIDs
}

// GetChatEventLog invokes method getChatEventLog#1e11b897 returning error if any.
func (c *Client) GetChatEventLog(ctx context.Context, request *GetChatEventLogRequest) (*ChatEvents, error) {
	var result ChatEvents

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}