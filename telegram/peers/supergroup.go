package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Supergroup is a supergroup Channel.
type Supergroup struct {
	Channel
}

// SlowmodeEnabled whether slow mode is enabled for groups to prevent flood in chat.
func (c Supergroup) SlowmodeEnabled() bool {
	return c.raw.GetSlowmodeEnabled()
}

// DisableSlowMode disables slow mode.
func (c Supergroup) DisableSlowMode(ctx context.Context) error {
	return c.ToggleSlowMode(ctx, -1)
}

// ToggleSlowMode Toggle supergroup slow mode: if enabled, users will only be able to send one message per seconds.
//
// If seconds is zero or smaller, slow mode will be disabled.
func (c Supergroup) ToggleSlowMode(ctx context.Context, seconds int) error {
	if seconds < 0 {
		seconds = 0
	}
	if _, err := c.m.api.ChannelsToggleSlowMode(ctx, &tg.ChannelsToggleSlowModeRequest{
		Channel: c.InputChannel(),
		Seconds: seconds,
	}); err != nil {
		return errors.Wrap(err, "toggle slow mode")
	}

	return nil
}

// SetStickerSet associates a sticker set to this supergroup.
func (c Supergroup) SetStickerSet(ctx context.Context, set tg.InputStickerSetClass) error {
	if _, err := c.m.api.ChannelsSetStickers(ctx, &tg.ChannelsSetStickersRequest{
		Channel:    c.InputChannel(),
		Stickerset: set,
	}); err != nil {
		return errors.Wrap(err, "set stickers")
	}

	return nil
}

// ResetStickerSet resets associated sticker set of this supergroup.
func (c Supergroup) ResetStickerSet(ctx context.Context) error {
	return c.SetStickerSet(ctx, &tg.InputStickerSetEmpty{})
}
