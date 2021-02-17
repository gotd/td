package message

import (
	"strings"
	"unicode/utf8"

	"github.com/gotd/td/tg"
)

type formatter func(offset, limit int) tg.MessageEntityClass

func appendMessage(entities []tg.MessageEntityClass, s string, format formatter) []tg.MessageEntityClass {
	length := utf8.RuneCountInString(strings.TrimSpace(s))
	return append(entities, format(0, length))
}

// AppendMention formats message as Mention message entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func AppendMention(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityMention{Offset: offset, Length: limit}
	})
}

// AppendHashtag formats message as Hashtag message entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func AppendHashtag(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityHashtag{Offset: offset, Length: limit}
	})
}

// AppendBotCommand formats message as BotCommand message entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func AppendBotCommand(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBotCommand{Offset: offset, Length: limit}
	})
}

// AppendURL formats message as Url message entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func AppendURL(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityUrl{Offset: offset, Length: limit}
	})
}

// AppendEmail formats message as Email message entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func AppendEmail(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityEmail{Offset: offset, Length: limit}
	})
}

// AppendBold formats message as Bold message entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func AppendBold(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBold{Offset: offset, Length: limit}
	})
}

// AppendItalic formats message as Italic message entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func AppendItalic(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityItalic{Offset: offset, Length: limit}
	})
}

// AppendCode formats message as Code message entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func AppendCode(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCode{Offset: offset, Length: limit}
	})
}

// AppendPre formats message as Pre message entity.
// See https://core.telegram.org/constructor/messageEntityPre.
func AppendPre(entities []tg.MessageEntityClass, s, lang string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPre{Offset: offset, Length: limit, Language: lang}
	})
}

// AppendTextURL formats message as TextUrl message entity.
// See https://core.telegram.org/constructor/messageEntityTextUrl.
func AppendTextURL(entities []tg.MessageEntityClass, s, url string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityTextUrl{Offset: offset, Length: limit, URL: url}
	})
}

// AppendMentionName formats message as MentionName message entity.
// See https://core.telegram.org/constructor/messageEntityMentionName.
func AppendMentionName(entities []tg.MessageEntityClass, s string, userID int) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityMentionName{Offset: offset, Length: limit, UserID: userID}
	})
}

// AppendPhone formats message as Phone message entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func AppendPhone(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPhone{Offset: offset, Length: limit}
	})
}

// AppendCashtag formats message as Cashtag message entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func AppendCashtag(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCashtag{Offset: offset, Length: limit}
	})
}

// AppendUnderline formats message as Underline message entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func AppendUnderline(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityUnderline{Offset: offset, Length: limit}
	})
}

// AppendStrike formats message as Strike message entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func AppendStrike(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityStrike{Offset: offset, Length: limit}
	})
}

// AppendBlockquote formats message as Blockquote message entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func AppendBlockquote(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBlockquote{Offset: offset, Length: limit}
	})
}

// AppendBankCard formats message as formats message entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func AppendBankCard(entities []tg.MessageEntityClass, s string) []tg.MessageEntityClass {
	return appendMessage(entities, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBankCard{Offset: offset, Length: limit}
	})
}

// FormatMention formats message as Mention message entity.
// See https://core.telegram.org/constructor/messageEntityMention.
func FormatMention(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityMention{Offset: offset, Length: limit}
	})
}

// FormatHashtag formats message as Hashtag message entity.
// See https://core.telegram.org/constructor/messageEntityHashtag.
func FormatHashtag(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityHashtag{Offset: offset, Length: limit}
	})
}

// FormatBotCommand formats message as BotCommand message entity.
// See https://core.telegram.org/constructor/messageEntityBotCommand.
func FormatBotCommand(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBotCommand{Offset: offset, Length: limit}
	})
}

// FormatURL formats message as Url message entity.
// See https://core.telegram.org/constructor/messageEntityUrl.
func FormatURL(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityUrl{Offset: offset, Length: limit}
	})
}

// FormatEmail formats message as Email message entity.
// See https://core.telegram.org/constructor/messageEntityEmail.
func FormatEmail(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityEmail{Offset: offset, Length: limit}
	})
}

// FormatBold formats message as Bold message entity.
// See https://core.telegram.org/constructor/messageEntityBold.
func FormatBold(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBold{Offset: offset, Length: limit}
	})
}

// FormatItalic formats message as Italic message entity.
// See https://core.telegram.org/constructor/messageEntityItalic.
func FormatItalic(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityItalic{Offset: offset, Length: limit}
	})
}

// FormatCode formats message as Code message entity.
// See https://core.telegram.org/constructor/messageEntityCode.
func FormatCode(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCode{Offset: offset, Length: limit}
	})
}

// FormatPre formats message as Pre message entity.
// See https://core.telegram.org/constructor/messageEntityPre.
func FormatPre(s, lang string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPre{Offset: offset, Length: limit, Language: lang}
	})
}

// FormatTextURL formats message as TextUrl message entity.
// See https://core.telegram.org/constructor/messageEntityTextUrl.
func FormatTextURL(s, url string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityTextUrl{Offset: offset, Length: limit, URL: url}
	})
}

// FormatMentionName formats message as MentionName message entity.
// See https://core.telegram.org/constructor/messageEntityMentionName.
func FormatMentionName(s string, userID int) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityMentionName{Offset: offset, Length: limit, UserID: userID}
	})
}

// FormatPhone formats message as Phone message entity.
// See https://core.telegram.org/constructor/messageEntityPhone.
func FormatPhone(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityPhone{Offset: offset, Length: limit}
	})
}

// FormatCashtag formats message as Cashtag message entity.
// See https://core.telegram.org/constructor/messageEntityCashtag.
func FormatCashtag(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityCashtag{Offset: offset, Length: limit}
	})
}

// FormatUnderline formats message as Underline message entity.
// See https://core.telegram.org/constructor/messageEntityUnderline.
func FormatUnderline(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityUnderline{Offset: offset, Length: limit}
	})
}

// FormatStrike formats message as Strike message entity.
// See https://core.telegram.org/constructor/messageEntityStrike.
func FormatStrike(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityStrike{Offset: offset, Length: limit}
	})
}

// FormatBlockquote formats message as Blockquote message entity.
// See https://core.telegram.org/constructor/messageEntityBlockquote.
func FormatBlockquote(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBlockquote{Offset: offset, Length: limit}
	})
}

// FormatBankCard formats message as formats message entity.
// See https://core.telegram.org/constructor/messageEntityBankCard.
func FormatBankCard(s string) []tg.MessageEntityClass {
	return appendMessage(nil, s, func(offset, limit int) tg.MessageEntityClass {
		return &tg.MessageEntityBankCard{Offset: offset, Length: limit}
	})
}
