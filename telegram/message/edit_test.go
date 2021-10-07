package message

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func TestEditMessageBuilder_Text(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer:    &tg.InputPeerSelf{},
		ID:      10,
		Message: msg,
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Edit(10).Text(ctx, msg)
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer:    &tg.InputPeerSelf{},
		ID:      10,
		Message: msg,
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().Edit(10).Textf(ctx, "%s", msg)
	require.Error(t, err)
}

func TestEditMessageBuilder_StyledText(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer:    &tg.InputPeerSelf{},
		ID:      10,
		Message: msg,
		Entities: []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Length: utf8.RuneCountInString(msg),
			},
		},
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Edit(10).StyledText(ctx, styling.Bold(msg))
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer:    &tg.InputPeerSelf{},
		ID:      10,
		Message: msg,
		Entities: []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Length: utf8.RuneCountInString(msg),
			},
		},
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().Edit(10).StyledText(ctx, styling.Bold(msg))
	require.Error(t, err)
}

func TestEditMessageBuilder_Media(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputPhoto{
		ID: 10,
	}

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   10,
		Media: &tg.InputMediaPhoto{
			ID: loc,
		},
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().Edit(10).Media(ctx, Photo(loc))
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   10,
		Media: &tg.InputMediaPhoto{
			ID: loc,
		},
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().Edit(10).Media(ctx, Photo(loc))
	require.Error(t, err)
}
