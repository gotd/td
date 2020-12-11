// Code generated by gotdgen, DO NOT EDIT.

package tg

import (
	"context"
	"fmt"

	"github.com/gotd/td/bin"
)

// No-op definition for keeping imports.
var _ = bin.Buffer{}
var _ = context.Background()
var _ = fmt.Stringer(nil)

// MessagesGetAttachedStickersRequest represents TL type `messages.getAttachedStickers#cc5b67cc`.
// Get stickers attached to a photo or video
//
// See https://core.telegram.org/method/messages.getAttachedStickers for reference.
type MessagesGetAttachedStickersRequest struct {
	// Stickered media
	Media InputStickeredMediaClass
}

// MessagesGetAttachedStickersRequestTypeID is TL type id of MessagesGetAttachedStickersRequest.
const MessagesGetAttachedStickersRequestTypeID = 0xcc5b67cc

// Encode implements bin.Encoder.
func (g *MessagesGetAttachedStickersRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode messages.getAttachedStickers#cc5b67cc as nil")
	}
	b.PutID(MessagesGetAttachedStickersRequestTypeID)
	if g.Media == nil {
		return fmt.Errorf("unable to encode messages.getAttachedStickers#cc5b67cc: field media is nil")
	}
	if err := g.Media.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.getAttachedStickers#cc5b67cc: field media: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *MessagesGetAttachedStickersRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode messages.getAttachedStickers#cc5b67cc to nil")
	}
	if err := b.ConsumeID(MessagesGetAttachedStickersRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.getAttachedStickers#cc5b67cc: %w", err)
	}
	{
		value, err := DecodeInputStickeredMedia(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.getAttachedStickers#cc5b67cc: field media: %w", err)
		}
		g.Media = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesGetAttachedStickersRequest.
var (
	_ bin.Encoder = &MessagesGetAttachedStickersRequest{}
	_ bin.Decoder = &MessagesGetAttachedStickersRequest{}
)

// MessagesGetAttachedStickers invokes method messages.getAttachedStickers#cc5b67cc returning error if any.
// Get stickers attached to a photo or video
//
// See https://core.telegram.org/method/messages.getAttachedStickers for reference.
func (c *Client) MessagesGetAttachedStickers(ctx context.Context, request *MessagesGetAttachedStickersRequest) ([]StickerSetCoveredClass, error) {
	var result StickerSetCoveredClassVector
	if err := c.rpc.InvokeRaw(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.Elems, nil
}
