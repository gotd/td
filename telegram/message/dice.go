package message

import "github.com/gotd/td/tg"

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
