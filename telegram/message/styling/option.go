package styling

import (
	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/tg"
)

// StyledTextOption is an option for styling text.
type StyledTextOption struct {
	size    int
	perform func(b *textBuilder) error
}

// Zero returns true if option is zero value.
func (s StyledTextOption) Zero() bool {
	return s.perform == nil
}

func styledTextOption(s string, perform func(b *textBuilder) error) StyledTextOption {
	return StyledTextOption{
		perform: perform,
		size:    len(s),
	}
}

// Custom formats text using given callback.
func Custom(cb func(eb *entity.Builder) error) StyledTextOption {
	return StyledTextOption{
		size: 0,
		perform: func(b *textBuilder) error {
			return cb(b.Builder)
		},
	}
}

// Plain formats text without any entities.
func Plain(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Plain(s)
		return nil
	})
}

// Mention formats text as Mention entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func Mention(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Mention(s)
		return nil
	})
}

// Hashtag formats text as Hashtag entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func Hashtag(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Hashtag(s)
		return nil
	})
}

// BotCommand formats text as BotCommand entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func BotCommand(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.BotCommand(s)
		return nil
	})
}

// URL formats text as URL entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func URL(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.URL(s)
		return nil
	})
}

// Email formats text as Email entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func Email(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Email(s)
		return nil
	})
}

// Bold formats text as Bold entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func Bold(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Bold(s)
		return nil
	})
}

// Italic formats text as Italic entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func Italic(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Italic(s)
		return nil
	})
}

// Code formats text as Code entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func Code(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Code(s)
		return nil
	})
}

// Pre formats text as Pre entity (without language).
// See https://core.telegram.org/constructor/messageEntityPre.
//
// Use PreLang to pass language.
func Pre(s string) StyledTextOption {
	return PreLang(s, "")
}

// PreLang formats text as Pre entity with language.
// See https://core.telegram.org/constructor/messageEntityPre.
func PreLang(s, lang string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Pre(s, lang)
		return nil
	})
}

// TextURL formats text as TextURL entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func TextURL(s, url string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.TextURL(s, url)
		return nil
	})
}

// MentionName formats text as MentionName entity.
// See https://core.telegram.org/constructor/messageEntityMentionName.
func MentionName(s string, userID tg.InputUserClass) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.MentionName(s, userID)
		return nil
	})
}

// Phone formats text as Phone entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func Phone(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Phone(s)
		return nil
	})
}

// Cashtag formats text as Cashtag entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func Cashtag(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Cashtag(s)
		return nil
	})
}

// Underline formats text as Underline entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func Underline(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Underline(s)
		return nil
	})
}

// Strike formats text as Strike entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func Strike(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Strike(s)
		return nil
	})
}

// Blockquote formats text as Blockquote entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func Blockquote(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Blockquote(s)
		return nil
	})
}

// BankCard formats text as BankCard entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func BankCard(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.BankCard(s)
		return nil
	})
}
