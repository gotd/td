package message

import "github.com/gotd/td/telegram/message/entity"

type textBuilder struct {
	entity.Builder
}

func (b *textBuilder) Perform(text StyledTextOption, texts ...StyledTextOption) {
	text(b)

	if len(texts) > 0 {
		for _, opt := range texts {
			opt(b)
		}
	}
}

// StyledTextOption is functional option for styling text.
type StyledTextOption func(builder *textBuilder)

// Plain formats text without any entities.
func Plain(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Plain(s)
	}
}

// Mention formats text as Mention entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func Mention(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Mention(s)
	}
}

// Hashtag formats text as Hashtag entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func Hashtag(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Hashtag(s)
	}
}

// BotCommand formats text as BotCommand entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func BotCommand(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.BotCommand(s)
	}
}

// URL formats text as URL entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func URL(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.URL(s)
	}
}

// Email formats text as Email entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func Email(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Email(s)
	}
}

// Bold formats text as Bold entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func Bold(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Bold(s)
	}
}

// Italic formats text as Italic entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func Italic(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Italic(s)
	}
}

// Code formats text as Code entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func Code(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Code(s)
	}
}

// Pre formats text as Pre entity.
// See https://core.telegram.org/constructor/messageEntityPre.
func Pre(s, lang string) StyledTextOption {
	return func(b *textBuilder) {
		b.Pre(s, lang)
	}
}

// TextURL formats text as TextURL entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func TextURL(s, url string) StyledTextOption {
	return func(b *textBuilder) {
		b.TextURL(s, url)
	}
}

// MentionName formats text as MentionName entity.
// See https://core.telegram.org/constructor/messageEntityMentionName.
func MentionName(s string, userID int) StyledTextOption {
	return func(b *textBuilder) {
		b.MentionName(s, userID)
	}
}

// Phone formats text as Phone entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func Phone(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Phone(s)
	}
}

// Cashtag formats text as Cashtag entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func Cashtag(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Cashtag(s)
	}
}

// Underline formats text as Underline entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func Underline(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Underline(s)
	}
}

// Strike formats text as Strike entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func Strike(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Strike(s)
	}
}

// Blockquote formats text as Blockquote entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func Blockquote(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.Blockquote(s)
	}
}

// BankCard formats text as BankCard entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func BankCard(s string) StyledTextOption {
	return func(b *textBuilder) {
		b.BankCard(s)
	}
}
