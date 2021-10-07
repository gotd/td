package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func TestPoll(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id := int64(0)
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)

		m, ok := req.Media.(*tg.InputMediaPoll)
		require.True(t, ok)
		id = m.Poll.ID
		require.Len(t, m.Poll.Answers, 3)
		require.Len(t, m.CorrectAnswers, 1)
		require.Equal(t, m.Poll.Answers[0].Option, m.CorrectAnswers[0])
	}).ThenResult(&tg.Updates{})
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)

		m, ok := req.Media.(*tg.InputMediaPoll)
		require.True(t, ok)
		require.Equal(t, id, m.Poll.ID)
		require.True(t, m.Poll.Closed)
	}).ThenResult(&tg.Updates{})

	poll := Poll("Nu che tam s den'gami?",
		CorrectPollAnswer("A?"),
		PollAnswer("Che?"),
		PollAnswer("Kuda?"),
	).PublicVoters(true).
		StyledExplanation(
			styling.Plain("See"),
			styling.TextURL("https://youtu.be/PYzX7SDKhd0.", "https://youtu.be/PYzX7SDKhd0"),
		)

	_, err := sender.Self().Media(ctx, poll)
	require.NoError(t, err)
	_, err = sender.Self().Media(ctx, poll.Close())
	require.NoError(t, err)
}

func TestRequestBuilder_PollVote(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesSendVoteRequest{
		Peer:    &tg.InputPeerSelf{},
		MsgID:   10,
		Options: [][]byte{[]byte("abc")},
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().PollVote(ctx, 10, []byte("abc"))
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesSendVoteRequest{
		Peer:    &tg.InputPeerSelf{},
		MsgID:   10,
		Options: [][]byte{[]byte("abc")},
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().PollVote(ctx, 10, []byte("abc"))
	require.Error(t, err)
}
