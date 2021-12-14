package peers

import (
	"context"

	"github.com/gotd/td/tg"
)

// Bot is a bot User.
type Bot struct {
	User
}

// ChatHistory whether can the bot see all messages in groups.
func (b Bot) ChatHistory() bool {
	return b.raw.GetBotChatHistory()
}

// CanBeAdded whether can the bot be added to group.
func (b Bot) CanBeAdded() bool {
	return !b.raw.GetBotNochats()
}

// InlineGeo whether the bot can request our geolocation in inline mode.
func (b Bot) InlineGeo() bool {
	return b.raw.GetBotInlineGeo()
}

// InlinePlaceholder returns inline placeholder for this inline bot.
func (b Bot) InlinePlaceholder() (string, bool) {
	return b.raw.GetBotInlinePlaceholder()
}

// SupportsInline whether the bot supports inline queries.
func (b Bot) SupportsInline() bool {
	_, ok := b.InlinePlaceholder()
	return ok
}

// BotInfo returns bot info.
func (b Bot) BotInfo(ctx context.Context) (tg.BotInfo, error) {
	full, err := b.m.getUserFull(ctx, b.InputUser())
	if err != nil {
		return tg.BotInfo{}, err
	}

	return full.BotInfo, nil
}
