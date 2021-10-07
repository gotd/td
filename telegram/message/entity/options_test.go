package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestBuilder(t *testing.T) {
	b := Builder{}
	t.Run("Plain", func(t *testing.T) {
		_, ent := b.Plain("abc").Complete()
		require.Empty(t, ent)
	})
	t.Run("Format", func(t *testing.T) {
		_, ent := b.Format("abc", Bold(), Italic()).Complete()
		require.Len(t, ent, 2)
		require.Equal(t, &tg.MessageEntityBold{
			Offset: 0,
			Length: len("abc"),
		}, ent[0])
		require.Equal(t, &tg.MessageEntityItalic{
			Offset: 0,
			Length: len("abc"),
		}, ent[1])
	})
	t.Run("Mention", func(t *testing.T) {
		_, ent := b.Mention("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityMention{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Hashtag", func(t *testing.T) {
		_, ent := b.Hashtag("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityHashtag{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("BotCommand", func(t *testing.T) {
		_, ent := b.BotCommand("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityBotCommand{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("URL", func(t *testing.T) {
		_, ent := b.URL("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityURL{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Email", func(t *testing.T) {
		_, ent := b.Email("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityEmail{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Bold", func(t *testing.T) {
		_, ent := b.Bold("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityBold{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Italic", func(t *testing.T) {
		_, ent := b.Italic("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityItalic{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Code", func(t *testing.T) {
		_, ent := b.Code("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityCode{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Pre", func(t *testing.T) {
		_, ent := b.Pre("abc", "lang").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityPre{
			Offset:   0,
			Length:   len("abc"),
			Language: "lang",
		}, r)
	})
	t.Run("TextURL", func(t *testing.T) {
		_, ent := b.TextURL("abc", "url").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityTextURL{
			Offset: 0,
			Length: len("abc"),
			URL:    "url",
		}, r)
	})
	t.Run("MentionName", func(t *testing.T) {
		user := &tg.InputUser{
			UserID:     10,
			AccessHash: 10,
		}
		_, ent := b.MentionName("abc", user).Complete()
		r := ent[0]
		require.Equal(t, &tg.InputMessageEntityMentionName{
			Offset: 0,
			Length: len("abc"),
			UserID: user,
		}, r)
	})
	t.Run("Phone", func(t *testing.T) {
		_, ent := b.Phone("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityPhone{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Cashtag", func(t *testing.T) {
		_, ent := b.Cashtag("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityCashtag{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Underline", func(t *testing.T) {
		_, ent := b.Underline("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityUnderline{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Strike", func(t *testing.T) {
		_, ent := b.Strike("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityStrike{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("Blockquote", func(t *testing.T) {
		_, ent := b.Blockquote("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityBlockquote{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
	t.Run("BankCard", func(t *testing.T) {
		_, ent := b.BankCard("abc").Complete()
		r := ent[0]
		require.Equal(t, &tg.MessageEntityBankCard{
			Offset: 0,
			Length: len("abc"),
		}, r)
	})
}
