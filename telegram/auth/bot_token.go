package auth

import (
	"strconv"
	"strings"

	"github.com/go-faster/errors"
)

// ErrInvalidBotToken is returned by ParseBotToken for a malformed bot token.
var ErrInvalidBotToken = errors.New("invalid bot token")

// maxBotTokenLength is the maximum length of a bot token accepted by the Bot API
// server.
const maxBotTokenLength = 80

// BotToken is a parsed bot token from @BotFather, "<bot id>:<secret>".
type BotToken struct {
	// ID is the user ID of the bot.
	ID int64
	// Secret is the part of the token after the colon.
	Secret string
}

// String returns the token in its original form.
func (t BotToken) String() string {
	return strconv.FormatInt(t.ID, 10) + ":" + t.Secret
}

// ParseBotToken parses and validates a bot token with the rules of the Bot API
// server: at most 80 characters, no '/', a decimal bot ID without leading zero
// in (0, 2^54), a colon and a non-empty secret.
//
// The returned error wraps ErrInvalidBotToken and never contains the secret.
func ParseBotToken(s string) (BotToken, error) {
	if len(s) > maxBotTokenLength {
		return BotToken{}, errors.Wrapf(ErrInvalidBotToken, "longer than %d characters", maxBotTokenLength)
	}
	if strings.Contains(s, "/") {
		return BotToken{}, errors.Wrap(ErrInvalidBotToken, "contains '/'")
	}
	rawID, secret, ok := strings.Cut(s, ":")
	if !ok {
		return BotToken{}, errors.Wrap(ErrInvalidBotToken, "no ':'")
	}
	if rawID == "" || rawID[0] == '0' {
		return BotToken{}, errors.Wrap(ErrInvalidBotToken, "bot ID is empty or has a leading zero")
	}
	for _, c := range rawID {
		if c < '0' || c > '9' {
			return BotToken{}, errors.Wrap(ErrInvalidBotToken, "bot ID is not a decimal number")
		}
	}
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 || id >= 1<<54 {
		return BotToken{}, errors.Wrap(ErrInvalidBotToken, "bot ID out of range")
	}
	if secret == "" {
		return BotToken{}, errors.Wrap(ErrInvalidBotToken, "empty secret")
	}
	return BotToken{ID: id, Secret: secret}, nil
}
