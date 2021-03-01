package message

import (
	"context"

	"github.com/gotd/td/tg"
)

// MediaDice adds a dice-based animated sticker.
func MediaDice(emoticon string) MediaOption {
	return Media(&tg.InputMediaDice{
		Emoticon: emoticon,
	})
}

// Dice adds a dice animated sticker.
func Dice() MediaOption {
	return MediaDice("🎲")
}

// Darts adds a darts animated sticker.
func Darts() MediaOption {
	return MediaDice("🎯")
}

// Basketball adds a basketball animated sticker.
func Basketball() MediaOption {
	return MediaDice("🏀")
}

// Dice sends a dice animated sticker.
func (b *Builder) Dice(ctx context.Context) error {
	return b.Media(ctx, Dice())
}

// Darts sends a darts animated sticker.
func (b *Builder) Darts(ctx context.Context) error {
	return b.Media(ctx, Darts())
}

// Basketball sends a basketball animated sticker.
func (b *Builder) Basketball(ctx context.Context) error {
	return b.Media(ctx, Basketball())
}
