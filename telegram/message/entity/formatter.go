package entity

import "github.com/nnqq/td/tg"

// Formatter is a message entity constructor.
type Formatter func(offset, limit int) tg.MessageEntityClass

// Mention formats message as Mention message entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func Mention() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityMention{Offset: offset, Length: limit}
	}
}

// Hashtag formats message as Hashtag message entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func Hashtag() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityHashtag{Offset: offset, Length: limit}
	}
}

// BotCommand formats message as BotCommand message entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func BotCommand() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBotCommand{Offset: offset, Length: limit}
	}
}

// URL formats message as Url message entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func URL() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityURL{Offset: offset, Length: limit}
	}
}

// Email formats message as Email message entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func Email() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityEmail{Offset: offset, Length: limit}
	}
}

// Bold formats message as Bold message entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func Bold() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBold{Offset: offset, Length: limit}
	}
}

// Italic formats message as Italic message entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func Italic() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityItalic{Offset: offset, Length: limit}
	}
}

// Code formats message as Code message entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func Code() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCode{Offset: offset, Length: limit}
	}
}

// Pre formats message as Pre message entity.
// See https://core.telegram.org/constructor/messageEntityPre.
func Pre(lang string) Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPre{Offset: offset, Length: limit, Language: lang}
	}
}

// TextURL formats message as TextUrl message entity.
// See https://core.telegram.org/constructor/messageEntityTextUrl.
func TextURL(url string) Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityTextURL{Offset: offset, Length: limit, URL: url}
	}
}

// MentionName formats message as MentionName message entity.
// See https://core.telegram.org/constructor/inputMessageEntityMentionName.
func MentionName(userID tg.InputUserClass) Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.InputMessageEntityMentionName{Offset: offset, Length: limit, UserID: userID}
	}
}

// Phone formats message as Phone message entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func Phone() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPhone{Offset: offset, Length: limit}
	}
}

// Cashtag formats message as Cashtag message entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func Cashtag() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCashtag{Offset: offset, Length: limit}
	}
}

// Underline formats message as Underline message entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func Underline() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityUnderline{Offset: offset, Length: limit}
	}
}

// Strike formats message as Strike message entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func Strike() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityStrike{Offset: offset, Length: limit}
	}
}

// Blockquote formats message as Blockquote message entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func Blockquote() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBlockquote{Offset: offset, Length: limit}
	}
}

// BankCard formats message as formats message entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func BankCard() Formatter {
	return func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBankCard{Offset: offset, Length: limit}
	}
}
