package message

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestBuilder_Text(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(msg, req.Message)
	}).ThenResult(&tg.Updates{})

	mock.NoError(sender.Self().Text(ctx, msg))
}

func TestBuilder_StyledText(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	tests := []struct {
		name    string
		format  func(msg string) StyledTextOption
		creator func(o int) tg.MessageEntityClass
	}{
		{"Mention", Mention, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityMention{Length: o}
		}},
		{"Hashtag", Hashtag, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityHashtag{Length: o}
		}},
		{"BotCommand", BotCommand, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBotCommand{Length: o}
		}},
		{"URL", URL, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityUrl{Length: o}
		}},
		{"Email", Email, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityEmail{Length: o}
		}},
		{"Bold", Bold, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBold{Length: o}
		}},
		{"Italic", Italic, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityItalic{Length: o}
		}},
		{"Code", Code, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityCode{Length: o}
		}},
		{"Phone", Phone, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityPhone{Length: o}
		}},
		{"Cashtag", Cashtag, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityCashtag{Length: o}
		}},
		{"Underline", Underline, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityUnderline{Length: o}
		}},
		{"Strike", Strike, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityStrike{Length: o}
		}},
		{"Blockquote", Blockquote, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBlockquote{Length: o}
		}},
		{"BankCard", BankCard, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBankCard{Length: o}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := "abc"
			mock.ExpectFunc(func(b bin.Encoder) {
				req, ok := b.(*tg.MessagesSendMessageRequest)
				mock.True(ok)
				mock.Equal(&tg.InputPeerSelf{}, req.Peer)
				mock.Equal(msg, req.Message)

				mock.NotEmpty(len(req.Entities))
				mock.Equal(test.creator(utf8.RuneCountInString(msg)), req.Entities[0])
			}).ThenResult(&tg.Updates{})

			mock.NoError(sender.Self().StyledText(ctx, test.format(msg)))
		})
	}
}
