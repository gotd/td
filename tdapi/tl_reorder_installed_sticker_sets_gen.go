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

// ReorderInstalledStickerSetsRequest represents TL type `reorderInstalledStickerSets#4c37c303`.
type ReorderInstalledStickerSetsRequest struct {
	// Pass true to change the order of mask sticker sets; pass false to change the order of
	// ordinary sticker sets
	IsMasks bool
	// Identifiers of installed sticker sets in the new correct order
	StickerSetIDs []Int64
}

// ReorderInstalledStickerSetsRequestTypeID is TL type id of ReorderInstalledStickerSetsRequest.
const ReorderInstalledStickerSetsRequestTypeID = 0x4c37c303

// Ensuring interfaces in compile-time for ReorderInstalledStickerSetsRequest.
var (
	_ bin.Encoder     = &ReorderInstalledStickerSetsRequest{}
	_ bin.Decoder     = &ReorderInstalledStickerSetsRequest{}
	_ bin.BareEncoder = &ReorderInstalledStickerSetsRequest{}
	_ bin.BareDecoder = &ReorderInstalledStickerSetsRequest{}
)

func (r *ReorderInstalledStickerSetsRequest) Zero() bool {
	if r == nil {
		return true
	}
	if !(r.IsMasks == false) {
		return false
	}
	if !(r.StickerSetIDs == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (r *ReorderInstalledStickerSetsRequest) String() string {
	if r == nil {
		return "ReorderInstalledStickerSetsRequest(nil)"
	}
	type Alias ReorderInstalledStickerSetsRequest
	return fmt.Sprintf("ReorderInstalledStickerSetsRequest%+v", Alias(*r))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ReorderInstalledStickerSetsRequest) TypeID() uint32 {
	return ReorderInstalledStickerSetsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ReorderInstalledStickerSetsRequest) TypeName() string {
	return "reorderInstalledStickerSets"
}

// TypeInfo returns info about TL type.
func (r *ReorderInstalledStickerSetsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "reorderInstalledStickerSets",
		ID:   ReorderInstalledStickerSetsRequestTypeID,
	}
	if r == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "IsMasks",
			SchemaName: "is_masks",
		},
		{
			Name:       "StickerSetIDs",
			SchemaName: "sticker_set_ids",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (r *ReorderInstalledStickerSetsRequest) Encode(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't encode reorderInstalledStickerSets#4c37c303 as nil")
	}
	b.PutID(ReorderInstalledStickerSetsRequestTypeID)
	return r.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (r *ReorderInstalledStickerSetsRequest) EncodeBare(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't encode reorderInstalledStickerSets#4c37c303 as nil")
	}
	b.PutBool(r.IsMasks)
	b.PutInt(len(r.StickerSetIDs))
	for idx, v := range r.StickerSetIDs {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare reorderInstalledStickerSets#4c37c303: field sticker_set_ids element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (r *ReorderInstalledStickerSetsRequest) Decode(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't decode reorderInstalledStickerSets#4c37c303 to nil")
	}
	if err := b.ConsumeID(ReorderInstalledStickerSetsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode reorderInstalledStickerSets#4c37c303: %w", err)
	}
	return r.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (r *ReorderInstalledStickerSetsRequest) DecodeBare(b *bin.Buffer) error {
	if r == nil {
		return fmt.Errorf("can't decode reorderInstalledStickerSets#4c37c303 to nil")
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode reorderInstalledStickerSets#4c37c303: field is_masks: %w", err)
		}
		r.IsMasks = value
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode reorderInstalledStickerSets#4c37c303: field sticker_set_ids: %w", err)
		}

		if headerLen > 0 {
			r.StickerSetIDs = make([]Int64, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value Int64
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare reorderInstalledStickerSets#4c37c303: field sticker_set_ids: %w", err)
			}
			r.StickerSetIDs = append(r.StickerSetIDs, value)
		}
	}
	return nil
}

// GetIsMasks returns value of IsMasks field.
func (r *ReorderInstalledStickerSetsRequest) GetIsMasks() (value bool) {
	return r.IsMasks
}

// GetStickerSetIDs returns value of StickerSetIDs field.
func (r *ReorderInstalledStickerSetsRequest) GetStickerSetIDs() (value []Int64) {
	return r.StickerSetIDs
}

// ReorderInstalledStickerSets invokes method reorderInstalledStickerSets#4c37c303 returning error if any.
func (c *Client) ReorderInstalledStickerSets(ctx context.Context, request *ReorderInstalledStickerSetsRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}