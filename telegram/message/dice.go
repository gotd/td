package message

import (
	"context"

	"github.com/nnqq/td/tg"
)

// MediaDice adds a dice-based animated sticker.
func MediaDice(emoticon string) MediaOption {
	return Media(&tg.InputMediaDice{
		Emoticon: emoticon,
	})
}

const (
	// DiceEmoticon is an emoticon to send dice sticker.
	DiceEmoticon = "🎲"
	// DartsEmoticon is an emoticon to send darts sticker.
	DartsEmoticon = "🎯"
	// BasketballEmoticon is an emoticon to send basketball sticker.
	BasketballEmoticon = "🏀"
	// FootballEmoticon is an emoticon to send football sticker.
	FootballEmoticon = "⚽"
	// CasinoEmoticon is an emoticon to send casino sticker.
	CasinoEmoticon = "🎰"
	// BowlingEmoticon is an emoticon to send bowling sticker.
	BowlingEmoticon = "🎳"
)

// Dice adds a dice animated sticker.
func Dice() MediaOption {
	return MediaDice(DiceEmoticon)
}

// Darts adds a darts animated sticker.
func Darts() MediaOption {
	return MediaDice(DartsEmoticon)
}

// Basketball adds a basketball animated sticker.
func Basketball() MediaOption {
	return MediaDice(BasketballEmoticon)
}

// Football adds a football animated sticker.
func Football() MediaOption {
	return MediaDice(FootballEmoticon)
}

// Casino adds a casino animated sticker.
func Casino() MediaOption {
	return MediaDice(CasinoEmoticon)
}

// Bowling adds a bowling animated sticker.
func Bowling() MediaOption {
	return MediaDice(BowlingEmoticon)
}

// Dice sends a dice animated sticker.
func (b *Builder) Dice(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Dice())
}

// Darts sends a darts animated sticker.
func (b *Builder) Darts(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Darts())
}

// Basketball sends a basketball animated sticker.
func (b *Builder) Basketball(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Basketball())
}

// Football sends a football animated sticker.
func (b *Builder) Football(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Football())
}

// Casino sends a casino animated sticker.
func (b *Builder) Casino(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Casino())
}

// Bowling sends a bowling animated sticker.
func (b *Builder) Bowling(ctx context.Context) (tg.UpdatesClass, error) {
	return b.Media(ctx, Bowling())
}
