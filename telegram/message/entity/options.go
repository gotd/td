package entity

import "github.com/nnqq/td/tg"

// Plain formats message as plain text.
func (b *Builder) Plain(s string) *Builder {
	b.message.WriteString(s)
	b.lastFormatIndex = len(b.entities)
	return b
}

// Format formats message using given formatters.
func (b *Builder) Format(s string, formats ...Formatter) *Builder {
	return b.appendMessage(s, formats...)
}

// Mention formats message as Mention message entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func (b *Builder) Mention(s string) *Builder {
	return b.Format(s, Mention())
}

// Hashtag formats message as Hashtag message entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func (b *Builder) Hashtag(s string) *Builder {
	return b.Format(s, Hashtag())
}

// BotCommand formats message as BotCommand message entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func (b *Builder) BotCommand(s string) *Builder {
	return b.Format(s, BotCommand())
}

// URL formats message as Url message entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func (b *Builder) URL(s string) *Builder {
	return b.Format(s, URL())
}

// Email formats message as Email message entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func (b *Builder) Email(s string) *Builder {
	return b.Format(s, Email())
}

// Bold formats message as Bold message entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func (b *Builder) Bold(s string) *Builder {
	return b.Format(s, Bold())
}

// Italic formats message as Italic message entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func (b *Builder) Italic(s string) *Builder {
	return b.Format(s, Italic())
}

// Code formats message as Code message entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func (b *Builder) Code(s string) *Builder {
	return b.Format(s, Code())
}

// Pre formats message as Pre message entity.
// See https://core.telegram.org/constructor/messageEntityPre.
func (b *Builder) Pre(s, lang string) *Builder {
	return b.Format(s, Pre(lang))
}

// TextURL formats message as TextUrl message entity.
// See https://core.telegram.org/constructor/messageEntityTextUrl.
func (b *Builder) TextURL(s, url string) *Builder {
	return b.Format(s, TextURL(url))
}

// MentionName formats message as MentionName message entity.
// See https://core.telegram.org/constructor/inputMessageEntityMentionName.
func (b *Builder) MentionName(s string, userID tg.InputUserClass) *Builder {
	return b.Format(s, MentionName(userID))
}

// Phone formats message as Phone message entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func (b *Builder) Phone(s string) *Builder {
	return b.Format(s, Phone())
}

// Cashtag formats message as Cashtag message entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func (b *Builder) Cashtag(s string) *Builder {
	return b.Format(s, Cashtag())
}

// Underline formats message as Underline message entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func (b *Builder) Underline(s string) *Builder {
	return b.Format(s, Underline())
}

// Strike formats message as Strike message entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func (b *Builder) Strike(s string) *Builder {
	return b.Format(s, Strike())
}

// Blockquote formats message as Blockquote message entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func (b *Builder) Blockquote(s string) *Builder {
	return b.Format(s, Blockquote())
}

// BankCard formats message as formats message entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func (b *Builder) BankCard(s string) *Builder {
	return b.Format(s, BankCard())
}
