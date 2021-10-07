package message

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func TestBuilder_Text(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, msg, req.Message)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Textf(ctx, "%s", msg)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, msg, req.Message)
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().Textf(ctx, "%s", msg)
	require.Error(t, err)
}

func TestBuilder_StyledText(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	tests := []struct {
		name    string
		format  func(msg string) StyledTextOption
		creator func(o int) tg.MessageEntityClass
	}{
		{"Mention", styling.Mention, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityMention{Length: o}
		}},
		{"Hashtag", styling.Hashtag, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityHashtag{Length: o}
		}},
		{"BotCommand", styling.BotCommand, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBotCommand{Length: o}
		}},
		{"URL", styling.URL, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityURL{Length: o}
		}},
		{"Email", styling.Email, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityEmail{Length: o}
		}},
		{"Bold", styling.Bold, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBold{Length: o}
		}},
		{"Italic", styling.Italic, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityItalic{Length: o}
		}},
		{"Code", styling.Code, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityCode{Length: o}
		}},
		{"Phone", styling.Phone, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityPhone{Length: o}
		}},
		{"Cashtag", styling.Cashtag, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityCashtag{Length: o}
		}},
		{"Underline", styling.Underline, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityUnderline{Length: o}
		}},
		{"Strike", styling.Strike, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityStrike{Length: o}
		}},
		{"Blockquote", styling.Blockquote, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBlockquote{Length: o}
		}},
		{"BankCard", styling.BankCard, func(o int) tg.MessageEntityClass {
			return &tg.MessageEntityBankCard{Length: o}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := "abc"
			mock.ExpectFunc(func(b bin.Encoder) {
				req, ok := b.(*tg.MessagesSendMessageRequest)
				require.True(t, ok)
				require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
				require.Equal(t, msg, req.Message)

				require.NotZero(t, len(req.Entities))
				require.Equal(t, test.creator(utf8.RuneCountInString(msg)), req.Entities[0])
			}).ThenResult(&tg.Updates{})

			_, err := sender.Self().StyledText(ctx, test.format(msg))
			require.NoError(t, err)

			mock.ExpectFunc(func(b bin.Encoder) {
				req, ok := b.(*tg.MessagesSendMessageRequest)
				require.True(t, ok)
				require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
				require.Equal(t, msg, req.Message)

				require.NotZero(t, len(req.Entities))
				require.Equal(t, test.creator(utf8.RuneCountInString(msg)), req.Entities[0])
			}).ThenRPCErr(testRPCError())

			_, err = sender.Self().StyledText(ctx, test.format(msg))
			require.Error(t, err)
		})
	}
}
