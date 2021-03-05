package message

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/gotd/td/tg"
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

	mock.NoError(sender.Self().Edit(10).Text(ctx, msg))

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer:    &tg.InputPeerSelf{},
		ID:      10,
		Message: msg,
	}).ThenRPCErr(testRPCError())

	mock.Error(sender.Self().Edit(10).Text(ctx, msg))
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

	mock.NoError(sender.Self().Edit(10).StyledText(ctx, Bold(msg)))

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

	mock.Error(sender.Self().Edit(10).StyledText(ctx, Bold(msg)))
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

	mock.NoError(sender.Self().Edit(10).Media(ctx, Photo(loc)))

	mock.ExpectCall(&tg.MessagesEditMessageRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   10,
		Media: &tg.InputMediaPhoto{
			ID: loc,
		},
	}).ThenRPCErr(testRPCError())

	mock.Error(sender.Self().Edit(10).Media(ctx, Photo(loc)))
}
