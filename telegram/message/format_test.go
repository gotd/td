package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFormat(t *testing.T) {
	t.Run("Mention", func(t *testing.T) {
		r := FormatMention("abc")[0]
		require.Equal(t, &tg.MessageEntityMention{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Hashtag", func(t *testing.T) {
		r := FormatHashtag("abc")[0]
		require.Equal(t, &tg.MessageEntityHashtag{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("BotCommand", func(t *testing.T) {
		r := FormatBotCommand("abc")[0]
		require.Equal(t, &tg.MessageEntityBotCommand{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("URL", func(t *testing.T) {
		r := FormatURL("abc")[0]
		require.Equal(t, &tg.MessageEntityUrl{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Email", func(t *testing.T) {
		r := FormatEmail("abc")[0]
		require.Equal(t, &tg.MessageEntityEmail{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Bold", func(t *testing.T) {
		r := FormatBold("abc")[0]
		require.Equal(t, &tg.MessageEntityBold{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Italic", func(t *testing.T) {
		r := FormatItalic("abc")[0]
		require.Equal(t, &tg.MessageEntityItalic{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Code", func(t *testing.T) {
		r := FormatCode("abc")[0]
		require.Equal(t, &tg.MessageEntityCode{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Pre", func(t *testing.T) {
		r := FormatPre("abc", "lang")[0]
		require.Equal(t, &tg.MessageEntityPre{
			Offset:   0,
			Length:   len("abc"),
			Language: "lang",
		}, r)
	})
	t.Run("TextURL", func(t *testing.T) {
		r := FormatTextURL("abc", "url")[0]
		require.Equal(t, &tg.MessageEntityTextUrl{
			Offset: 0,
			Length: len("abc"),
			URL:    "url",
		}, r)
	})
	t.Run("MentionName", func(t *testing.T) {
		r := FormatMentionName("abc", 1)[0]
		require.Equal(t, &tg.MessageEntityMentionName{
			Offset: 0,
			Length: len("abc"),
			UserID: 1,
		}, r)
	})
	t.Run("Phone", func(t *testing.T) {
		r := FormatPhone("abc")[0]
		require.Equal(t, &tg.MessageEntityPhone{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Cashtag", func(t *testing.T) {
		r := FormatCashtag("abc")[0]
		require.Equal(t, &tg.MessageEntityCashtag{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Underline", func(t *testing.T) {
		r := FormatUnderline("abc")[0]
		require.Equal(t, &tg.MessageEntityUnderline{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Strike", func(t *testing.T) {
		r := FormatStrike("abc")[0]
		require.Equal(t, &tg.MessageEntityStrike{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Blockquote", func(t *testing.T) {
		r := FormatBlockquote("abc")[0]
		require.Equal(t, &tg.MessageEntityBlockquote{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("BankCard", func(t *testing.T) {
		r := FormatBankCard("abc")[0]
		require.Equal(t, &tg.MessageEntityBankCard{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
}
